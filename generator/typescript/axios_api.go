// Package typescript generates code for TS type definitions
// and http calls using the Axios library
package typescript

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/benoitkugler/gomacro/analysis"
	"github.com/benoitkugler/gomacro/analysis/httpapi"
	"github.com/benoitkugler/gomacro/generator"
)

// return arg: String(params[arg])
func asObjectKey(param httpapi.TypedParam) string {
	switch param.Type.Kind() {
	case analysis.BKFloat, analysis.BKInt:
		return fmt.Sprintf("%q: String(params[%q])", param.Name, param.Name) // stringify
	case analysis.BKBool:
		return fmt.Sprintf("%q: params[%q] ? 'ok' : ''", param.Name, param.Name) // stringify
	case analysis.BKString:
		return fmt.Sprintf("%q: params[%q]", param.Name, param.Name) // no converter
	default:
		panic(analysis.ExhaustiveAnonymousTypeSwitch)
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

func funcArgsName(a httpapi.Endpoint) string {
	if withFormData(a) { // form data mode
		if fi := a.Contract.InputForm.File; fi != "" {
			return "params, file"
		}
	} else if !hasBodyInput(a) {
		// params as query params
		if len(a.Contract.InputQueryParams) == 0 {
			return ""
		}
	}
	return "params"
}

func typeIn(a httpapi.Endpoint) string {
	if withFormData(a) { // form data mode
		params := "params: " + paramsType(a.Contract.InputForm.AsTypedValues())
		if fi := a.Contract.InputForm.File; fi != "" {
			params += ", file: File"
		}
		return params
	} else if hasBodyInput(a) { // JSON mode
		return "params: " + typeName(a.Contract.InputBody)
	}
	// params as query params
	if len(a.Contract.InputQueryParams) == 0 {
		return ""
	}
	return "params: " + paramsType(a.Contract.InputQueryParams)
}

// assume a named type as return value
func typeOut(a httpapi.Endpoint) string {
	return typeName(a.Contract.Return)
}

func fullUrl(a httpapi.Endpoint) string {
	return fmt.Sprintf("this.baseUrl + %q", a.Url) // basic url
}

func convertTypedQueryParams(c httpapi.Contract) string {
	chunks := make([]string, len(c.InputQueryParams))
	for i, param := range c.InputQueryParams {
		chunks[i] = asObjectKey(param)
	}
	return "{ " + strings.Join(chunks, ", ") + " }"
}

func generateAxiosCall(a httpapi.Endpoint) string {
	callParams := "{ headers: this.getHeaders() }"
	if len(a.Contract.InputQueryParams) != 0 {
		callParams = fmt.Sprintf("{ params: %s, headers : this.getHeaders() }", convertTypedQueryParams(a.Contract))
	}

	var template string
	if withFormData(a) { // add the creation of FormData
		template += "const formData = new FormData()\n"
		if fi := a.Contract.InputForm.File; fi != "" {
			template += fmt.Sprintf("formData.append(%q, file, file.name)\n", fi)
		}
		for _, param := range a.Contract.InputForm.ValueNames {
			template += fmt.Sprintf("formData.append(%q, params[%q])\n", param, param)
		}
		template += "const rep:AxiosResponse<%s> = await Axios.%s(fullUrl, formData, %s)"
	} else if hasBodyInput(a) {
		template = "const rep:AxiosResponse<%s> = await Axios.%s(fullUrl, params, %s)"
	} else if !hasBodyInput(a) && expectBodyParam(a) {
		template = "const rep:AxiosResponse<%s> = await Axios.%s(fullUrl, null, %s)"
	} else {
		template = "const rep:AxiosResponse<%s> = await Axios.%s(fullUrl, %s)"
	}

	methodLower := strings.ToLower(a.Method)

	return fmt.Sprintf(template, typeOut(a), methodLower, callParams)
}

func generateMethod(a httpapi.Endpoint) string {
	const template = `
	protected async raw%s(%s) {
		const fullUrl = %s;
		%s;
		return rep.data;
	}
	
	/** %s wraps raw%s and handles the error */
	async %s(%s) {
		this.startRequest();
		try {
			const out = await this.raw%s(%s);
			this.onSuccess%s(out);
			return out
		} catch (error) {
			this.handleError(error);
		}
	}

	protected abstract onSuccess%s(data: %s): void 
	`
	fnName := a.Contract.Name
	in := typeIn(a)
	return fmt.Sprintf(template,
		fnName, in, fullUrl(a), generateAxiosCall(a), fnName, fnName, fnName, in,
		fnName, funcArgsName(a), fnName, fnName, typeOut(a))
}

func renderTypes(s []httpapi.Endpoint) string {
	var (
		decls []generator.Declaration
		cache = make(generator.Cache)
	)
	for _, api := range s { // write top-level decl
		if ty := api.Contract.InputBody; ty != nil {
			decls = append(decls, generate(ty, cache)...)
		}
		if ty := api.Contract.Return; ty != nil {
			decls = append(decls, generate(ty, cache)...)
		}
	}
	return generator.WriteDeclarations(decls)
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
		constructor(protected baseUrl: string, protected authToken: string) {}

		abstract handleError(error: any): void

		abstract startRequest(): void

		getHeaders() {
			return { Authorization: "Bearer " + this.authToken }
		}

		%s
	}`, typesCode, strings.Join(apiCalls, "\n"))
}
