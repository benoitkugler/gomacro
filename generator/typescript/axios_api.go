// Package typescript generates code for TS type definitions
// and http calls using the Axios library
package typescript

import (
	"fmt"
	"net/http"
	"strings"

	an "github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/analysis/httpapi"
	"github.com/benoitkugler/gomacro/generator"
)

// return arg: String(params[arg])
func asObjectKey(param httpapi.TypedParam) string {
	var underlying *an.Basic
	switch t := param.Type.(type) {
	case *an.Basic:
		underlying = t
	case *an.Named:
		underlying = t.Underlying.(*an.Basic)
	default:
		panic("unsupported type")
	}

	switch underlying.Kind() {
	case an.BKFloat, an.BKInt:
		return fmt.Sprintf("%q: String(params[%q])", param.Name, param.Name) // stringify
	case an.BKBool:
		return fmt.Sprintf("%q: params[%q] ? 'ok' : ''", param.Name, param.Name) // stringify
	case an.BKString:
		return fmt.Sprintf("%q: params[%q]", param.Name, param.Name) // no converter
	default:
		panic(an.ExhaustiveAnonymousTypeSwitch)
	}
}

// returns true if the client API call has an argument
// for the body
func expectBodyParam(a httpapi.Endpoint) bool {
	return a.Method == http.MethodPost || a.Method == http.MethodPut
}

func hasBodyInput(a httpapi.Endpoint) bool {
	return a.Contract.InputBody != nil
}

func withFormData(a httpapi.Endpoint) bool {
	return !a.Contract.InputForm.IsZero()
}

func paramsType(params []httpapi.TypedParam) string {
	tmp := make([]string, len(params))
	for i, param := range params {
		tmp[i] = fmt.Sprintf("%q: %s", param.Name, typeName(param.Type)) // quote for names like "id-1"
	}
	return "{" + strings.Join(tmp, ", ") + "}"
}

func typeIn(a httpapi.Endpoint) string {
	if hasBodyInput(a) { // JSON mode
		return "params: " + typeName(a.Contract.InputBody)
	}

	var chunks []string
	if withFormData(a) { // form data mode
		if vals := a.Contract.InputForm.AsTypedValues(); len(vals) != 0 {
			chunks = append(chunks, "formParams: "+paramsType(vals))
		}
		if fi := a.Contract.InputForm.File; fi != "" {
			chunks = append(chunks, "file: File")
		}
		if json := a.Contract.InputForm.JSON; json.Name != "" {
			chunks = append(chunks, "formValue: "+typeName(json.Type))
		}
	}

	// params as query params
	if len(a.Contract.InputQueryParams) != 0 {
		chunks = append(chunks, "params: "+paramsType(a.Contract.InputQueryParams))
	}

	return strings.Join(chunks, ", ")
}

func hasNoReturn(a httpapi.Endpoint) bool { return a.Contract.Return == nil }

// assume a named type as return value
func typeOut(a httpapi.Endpoint) string {
	if a.Contract.IsReturnBlob {
		return "Blob"
	}
	if hasNoReturn(a) {
		return "never"
	}
	return typeName(a.Contract.Return)
}

func fullUrl(a httpapi.Endpoint) string {
	return fmt.Sprintf("this.baseURL + %q", a.Url) // basic url
}

func convertTypedQueryParams(c httpapi.Contract) string {
	chunks := make([]string, len(c.InputQueryParams))
	for i, param := range c.InputQueryParams {
		chunks[i] = asObjectKey(param)
	}
	return "{ " + strings.Join(chunks, ", ") + " }"
}

func generateAxiosCall(a httpapi.Endpoint) string {
	callObjectItems := []string{
		"headers: this.getHeaders()",
	}
	if len(a.Contract.InputQueryParams) != 0 {
		callObjectItems = append(callObjectItems, fmt.Sprintf("params: %s", convertTypedQueryParams(a.Contract)))
	}
	if a.Contract.IsReturnBlob {
		callObjectItems = append(callObjectItems, "responseType: 'arraybuffer'")
	}
	callParams := fmt.Sprintf("{ %s }", strings.Join(callObjectItems, ", "))

	returnAssignment := ""
	if !hasNoReturn(a) {
		returnAssignment = fmt.Sprintf("const rep:AxiosResponse<%s> = ", typeOut(a))
	}

	methodLower := strings.ToLower(a.Method)

	if withFormData(a) { // add the creation of FormData
		form := "const formData = new FormData()\n"
		if fi := a.Contract.InputForm.File; fi != "" {
			form += fmt.Sprintf("formData.append(%q, file, file.name)\n", fi)
		}
		for _, param := range a.Contract.InputForm.ValueNames {
			form += fmt.Sprintf("formData.append(%q, formParams[%q])\n", param, param)
		}
		if param := a.Contract.InputForm.JSON; param.Name != "" {
			form += fmt.Sprintf("formData.append(%q, JSON.stringify(formValue))\n", param.Name)
		}
		return fmt.Sprintf("%s %s await Axios.%s(fullUrl, formData, %s)", form, returnAssignment, methodLower, callParams)
	} else if hasBodyInput(a) {
		return fmt.Sprintf("%s await Axios.%s(fullUrl, params, %s)", returnAssignment, methodLower, callParams)
	} else if !hasBodyInput(a) && expectBodyParam(a) {
		return fmt.Sprintf("%s await Axios.%s(fullUrl, null, %s)", returnAssignment, methodLower, callParams)
	} else {
		return fmt.Sprintf("%s await Axios.%s(fullUrl, %s)", returnAssignment, methodLower, callParams)
	}
}

func generateMethodJSONStream(a httpapi.Endpoint) string {
	body := ""
	if hasBodyInput(a) {
		body = "body: JSON.stringify(params),"
	}
	const template = `
	/** %[1]s return a streaming Response (JSON line format) */
	async %[1]s(%[2]s) {
		const fullUrl = %[3]s;
		this.startRequest();
		try {
			const response = await fetch(fullUrl, {
				method: %[4]q,
				headers: {
					...this.getHeaders(),
					Accept: "application/json",
					"Content-Type": "application/json",
				},
			%[5]s
			});
			return response;
		} catch (error) {
			this.handleError(error);
		}
	}
`
	return fmt.Sprintf(template,
		a.Contract.Name, typeIn(a), fullUrl(a), a.Method, body)
}

func generateMethod(a httpapi.Endpoint) string {
	if a.Comment == httpapi.JSONStream { // defer to special fetch syntax
		return generateMethodJSONStream(a)
	}
	const template = `
	/** %[1]s performs the request and handles the error */
	async %[1]s(%[2]s) {
		const fullUrl = %[3]s;
		this.startRequest();
		try {
			%[4]s;
			%[5]s
		} catch (error) {
			this.handleError(error);
		}
	}
	`
	fnName := a.Contract.Name
	in := typeIn(a)
	returnValue := "return rep.data;"
	if hasNoReturn(a) {
		returnValue = "return true;"
	} else if a.Contract.IsReturnBlob {
		returnValue = `
		const header = rep.headers["content-disposition"]
		const startIndex = header.indexOf("filename=") + 9; 
		const endIndex = header.length; 
		const filename = decodeURIComponent(header.substring(startIndex, endIndex));
		return { blob: rep.data, filename: filename};
		`
	}
	return fmt.Sprintf(template,
		fnName, in, fullUrl(a),
		generateAxiosCall(a), returnValue)
}

func renderTypes(s []httpapi.Endpoint) string {
	var allTypes []an.Type
	for _, api := range s { // write top-level decl
		if ty := api.Contract.InputBody; ty != nil {
			allTypes = append(allTypes, ty)
		}
		if ty := api.Contract.Return; ty != nil {
			allTypes = append(allTypes, ty)
		}
		for _, qp := range api.Contract.InputQueryParams {
			allTypes = append(allTypes, qp.Type)
		}
	}
	return generator.WriteDeclarations(generateTypes(allTypes))
}

// GenerateAxios generate a TS class using Axios for calling the
// given http API description.
func GenerateAxios(api []httpapi.Endpoint) string {
	// generate the code required for all the endpoints
	typesCode := renderTypes(api)

	apiCalls := make([]string, len(api))
	for i, endpoint := range api {
		apiCalls[i] = generateMethod(endpoint)
	}

	return fmt.Sprintf(`
	// Code generated by gomacro/typescript/axios_api.go. DO NOT EDIT
	
	import type { AxiosResponse } from "axios";
	import Axios from "axios";

	%s

	/** AbstractAPI provides auto-generated API calls and should be used 
		as base class for an app controller.
	*/
	export abstract class AbstractAPI {
		constructor(protected baseURL: string, public authToken: string) {}

		abstract protected handleError(error: any): void

		abstract protected startRequest(): void

		getHeaders() {
			return { Authorization: "Bearer " + this.authToken }
		}

		%s
	}`, typesCode, strings.Join(apiCalls, "\n"))
}
