package dart

import (
	"fmt"
	"strings"

	an "github.com/benoitkugler/gomacro/analysis"
)

// generate serialization and deserialization for JSON format
// for named type we generate helpers function, unless
// they are basic

// jsonID returns the Dart prefix used when definiting a JSON function
// for this type
func jsonID(typ an.Type) string {
	switch typ := typ.(type) {
	case *an.Pointer:
		panic("pointers not handled by the Dart generator")
	case *an.Named: // directly call the underlying function
		switch typ.Underlying.(type) {
		case *an.Array, *an.Map:
			return lowerFirst(typeName(typ))
		default:
			return jsonID(typ.Underlying)
		}
	case *an.Basic, *an.Time:
		return lowerFirst(typeName(typ))
	case *an.Array:
		return "list" + strings.Title(jsonID(typ.Elem))
	case *an.Map:
		return "dict" + strings.Title(jsonID(typ.Key)) + "To" + strings.Title(jsonID(typ.Elem))
	case *an.Struct, *an.Enum, *an.Union: // these types are always named
		return lowerFirst(typeName(typ))
	default:
		panic(an.ExhaustiveTypeSwitch)
	}
}

func jsonForBasic(b *an.Basic) string {
	switch b.Kind() {
	case an.BKFloat: // use num to avoid issue with integers values
		return `double doubleFromJson(dynamic json) => (json as num).toDouble();
	
	double doubleToJson(double item) => item;
	
	`
	case an.BKString: // accept null
		return `String stringFromJson(dynamic json) => json == null ? "" : json as String;
		
		String stringToJson(String item) => item;
		
		`
	case an.BKBool, an.BKInt:
		name, id := typeName(b), jsonID(b)
		return fmt.Sprintf(`%s %sFromJson(dynamic json) => json as %s;
		
		%s %sToJson(%s item) => item;
		
		`, name, id, name, name, id, name)
	default:
		panic(an.ExhaustiveBasicKindSwitch)
	}
}

func jsonForTime() string {
	return `DateTime dateTimeFromJson(dynamic json) => DateTime.parse(json as String);

	dynamic dateTimeToJson(DateTime dt) => dt.toIso8601String();
	`
}

func jsonForEnum(en *an.Enum) string {
	valueType := "String"
	if en.IsInteger() {
		valueType = "int"
	}
	name, id := typeName(en), jsonID(en)

	return fmt.Sprintf(`%s %sFromJson(dynamic json) => _%sExt.fromValue(json as %s);
	
	dynamic %sToJson(%s item) => item.toValue();
	
	`, name, id, name, valueType, id, name)
}

func jsonForNamed(na *an.Named) string {
	switch na.Underlying.(type) {
	case *an.Array, *an.Map:
	default:
		return ""
	}
	name, id := typeName(na), jsonID(na)
	elemID := jsonID(na.Underlying)
	return fmt.Sprintf(`%s %sFromJson(dynamic json) { return %sFromJson(json); }

	dynamic %sToJson(%s item) { return %sToJson(item); }
	`, name, id, elemID,
		id, name, elemID)
}

func jsonForArray(l *an.Array) string {
	name, id := typeName(l), jsonID(l)
	elemID := jsonID(l.Elem)

	// nil slices are jsonized as null, check for it then
	return fmt.Sprintf(`%s %sFromJson(dynamic json) {
		if (json == null) {
			return [];
		}
		return (json as List<dynamic>).map(%sFromJson).toList();
	}

	List<dynamic> %sToJson(%s item) {
		return item.map(%sToJson).toList();
	}
	`, name, id, elemID, id, name, elemID)
}

func jsonForMap(ma *an.Map) string {
	keyName, keyID := typeName(ma.Key), jsonID(ma.Key)

	// JSON map keys are always string, but it is very convenient
	// to support int keys (for IDs)
	keyFromJson := "k as " + keyName
	if keyName == "int" {
		keyFromJson = "int.parse(k)"
	}

	name, id := typeName(ma), jsonID(ma)
	elemID := jsonID(ma.Elem)
	// nil dict are jsonized as null, check for it then
	return fmt.Sprintf(`%s %sFromJson(dynamic json) {
		if (json == null) {
			return {};
		}
		return (json as Map<String, dynamic>).map((k,v) => MapEntry(%s, %sFromJson(v)));
	}
	
	Map<String, dynamic> %sToJson(%s item) {
		return item.map((k,v) => MapEntry(%sToJson(k).toString(), %sToJson(v)));
	}
	`, name, id, keyFromJson, elemID, id, name, keyID, elemID)
}

func jsonForStruct(st *an.Struct) string {
	var fieldsFrom, fieldsTo []string
	for _, field := range st.Fields {
		if !field.JSONExported() {
			continue
		}

		fieldTypeID := jsonID(field.Type)
		fieldName := field.JSONName()
		dartFieldName := lowerFirst(fieldName) // convert to dart convention

		if field.IsOpaqueFor("dart") {
			// use identity
			fieldsFrom = append(fieldsFrom, fmt.Sprintf("json['%s']", fieldName))
			fieldsTo = append(fieldsTo, fmt.Sprintf("%q :  item.%s", fieldName, dartFieldName))
		} else {
			fieldsFrom = append(fieldsFrom, fmt.Sprintf("%sFromJson(json['%s'])", fieldTypeID, fieldName))
			fieldsTo = append(fieldsTo, fmt.Sprintf("%q : %sToJson(item.%s)", fieldName, fieldTypeID, dartFieldName))
		}
	}

	name, id := typeName(st), jsonID(st)

	return fmt.Sprintf(`
	%s %sFromJson(dynamic json_) {
		final json = (json_ as Map<String, dynamic>);
		return %s(
			%s
		);
	}
	
	Map<String, dynamic> %sToJson(%s item) {
		return {
			%s
		};
	}
	
	`, name, id, name, strings.Join(fieldsFrom, ",\n"),
		id, name, strings.Join(fieldsTo, ",\n"),
	)
}

func jsonForUnion(u *an.Union) string {
	var casesFrom, casesTo []string

	for i, member := range u.Members {
		kindTag := an.LocalName(member) // union members are always named

		memberName, memberID := typeName(member), jsonID(member)
		casesFrom = append(casesFrom, fmt.Sprintf(`case %q:
			return %sFromJson(data);`, kindTag, memberID))

		caseTo := fmt.Sprintf(`if (item is %s) {
			return {'Kind': %q, 'Data': %sToJson(item)};
		}`, memberName, kindTag, memberID)
		if i != 0 {
			caseTo = "else " + caseTo
		}
		casesTo = append(casesTo, caseTo)
	}

	name, id := typeName(u), jsonID(u)

	codeFrom := fmt.Sprintf(`%s %sFromJson(dynamic json_) {
		final json = json_ as Map<String, dynamic>;
		final kind = json['Kind'] as String;
		final data = json['Data'];
		switch (kind) {
			%s
		default:
			throw ("unexpected type");
		}
	}
	`, name, id, strings.Join(casesFrom, "\n"))

	codeTo := fmt.Sprintf(`Map<String, dynamic> %sToJson(%s item) {
		%s else {
			throw ("unexpected type");
		}	
	}
	`, id, name, strings.Join(casesTo, ""))

	return codeFrom + "\n" + codeTo
}
