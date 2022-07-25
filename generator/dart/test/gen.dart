// Code generated by gomacro/generator/dart. DO NOT EDIT

import 'extern2.dart';

// github.com/benoitkugler/gomacro/testutils/testsource.MyDate
typedef MyDate = DateTime;

String stringFromJson(dynamic json) => json == null ? "" : json as String;

String stringToJson(String item) => item;

// github.com/benoitkugler/gomacro/testutils/testsource/subpackage.StructWithComment
class StructWithComment {
  final int a;

  const StructWithComment(this.a);

  @override
  String toString() {
    return "StructWithComment($a)";
  }
}

StructWithComment structWithCommentFromJson(dynamic json_) {
  final json = (json_ as JSON);
  return StructWithComment(intFromJson(json['A']));
}

JSON structWithCommentToJson(StructWithComment item) {
  return {"A": intToJson(item.a)};
}

DateTime dateTimeFromJson(dynamic json) => DateTime.parse(json as String);

dynamic dateTimeToJson(DateTime dt) => dt.toString();

typedef JSON = Map<String, dynamic>; // alias to shorten JSON convertors

// github.com/benoitkugler/gomacro/testutils/testsource.basic1
typedef basic1 = int;

// github.com/benoitkugler/gomacro/testutils/testsource.basic2
typedef basic2 = bool;

// github.com/benoitkugler/gomacro/testutils/testsource.basic3
typedef basic3 = double;

// github.com/benoitkugler/gomacro/testutils/testsource.basic4
typedef basic4 = String;

bool boolFromJson(dynamic json) => json as bool;

bool boolToJson(bool item) => item;

// github.com/benoitkugler/gomacro/testutils/testsource.complexStruct
class complexStruct {
  final Map<int, int> dict;
  final int u;
  final DateTime time;
  final String b;
  final itfType value;
  final itfList l;
  final int a;
  final enumInt e;
  final enumUInt e2;
  final enumString e3;
  final MyDate date;
  final List<int> f;
  final StructWithComment imported;

  const complexStruct(this.dict, this.u, this.time, this.b, this.value, this.l,
      this.a, this.e, this.e2, this.e3, this.date, this.f, this.imported);

  @override
  String toString() {
    return "complexStruct($dict, $u, $time, $b, $value, $l, $a, $e, $e2, $e3, $date, $f, $imported)";
  }
}

complexStruct complexStructFromJson(dynamic json_) {
  final json = (json_ as JSON);
  return complexStruct(
      dictInt_IntFromJson(json['Dict']),
      intFromJson(json['U']),
      dateTimeFromJson(json['Time']),
      stringFromJson(json['B']),
      itfTypeFromJson(json['Value']),
      listItfTypeFromJson(json['L']),
      intFromJson(json['A']),
      enumIntFromJson(json['E']),
      enumUIntFromJson(json['E2']),
      enumStringFromJson(json['E3']),
      dateTimeFromJson(json['Date']),
      listIntFromJson(json['F']),
      structWithCommentFromJson(json['Imported']));
}

JSON complexStructToJson(complexStruct item) {
  return {
    "Dict": dictInt_IntToJson(item.dict),
    "U": intToJson(item.u),
    "Time": dateTimeToJson(item.time),
    "B": stringToJson(item.b),
    "Value": itfTypeToJson(item.value),
    "L": listItfTypeToJson(item.l),
    "A": intToJson(item.a),
    "E": enumIntToJson(item.e),
    "E2": enumUIntToJson(item.e2),
    "E3": enumStringToJson(item.e3),
    "Date": dateTimeToJson(item.date),
    "F": listIntToJson(item.f),
    "Imported": structWithCommentToJson(item.imported)
  };
}

// github.com/benoitkugler/gomacro/testutils/testsource.concretType1
class concretType1 implements itfType2, itfType {
  final List<int> list2;
  final int v;

  const concretType1(this.list2, this.v);

  @override
  String toString() {
    return "concretType1($list2, $v)";
  }
}

concretType1 concretType1FromJson(dynamic json_) {
  final json = (json_ as JSON);
  return concretType1(listIntFromJson(json['List2']), intFromJson(json['V']));
}

JSON concretType1ToJson(concretType1 item) {
  return {"List2": listIntToJson(item.list2), "V": intToJson(item.v)};
}

// github.com/benoitkugler/gomacro/testutils/testsource.concretType2
class concretType2 implements itfType {
  final double d;

  const concretType2(this.d);

  @override
  String toString() {
    return "concretType2($d)";
  }
}

concretType2 concretType2FromJson(dynamic json_) {
  final json = (json_ as JSON);
  return concretType2(doubleFromJson(json['D']));
}

JSON concretType2ToJson(concretType2 item) {
  return {"D": doubleToJson(item.d)};
}

Map<int, int> dictInt_IntFromJson(dynamic json) {
  if (json == null) {
    return {};
  }
  return (json as JSON).map((k, v) => MapEntry(int.parse(k), intFromJson(v)));
}

Map<String, dynamic> dictInt_IntToJson(Map<int, int> item) {
  return item.map((k, v) => MapEntry(intToJson(k).toString(), intToJson(v)));
}

double doubleFromJson(dynamic json) => (json as num).toDouble();

double doubleToJson(double item) => item;

// github.com/benoitkugler/gomacro/testutils/testsource.enumInt
enum enumInt { ai, bi, ci, di }

extension _enumIntExt on enumInt {
  static const _values = [0, 1, 2, 4];
  static enumInt fromValue(int s) {
    return enumInt.values[_values.indexOf(s)];
  }

  int toValue() {
    return _values[index];
  }
}

enumInt enumIntFromJson(dynamic json) => _enumIntExt.fromValue(json as int);

dynamic enumIntToJson(enumInt item) => item.toValue();

// github.com/benoitkugler/gomacro/testutils/testsource.enumString
enum enumString { sA, sB, sC, sD }

extension _enumStringExt on enumString {
  static const _values = ["va", "vb", "vc", "vd"];
  static enumString fromValue(String s) {
    return enumString.values[_values.indexOf(s)];
  }

  String toValue() {
    return _values[index];
  }
}

enumString enumStringFromJson(dynamic json) =>
    _enumStringExt.fromValue(json as String);

dynamic enumStringToJson(enumString item) => item.toValue();

// github.com/benoitkugler/gomacro/testutils/testsource.enumUInt
enum enumUInt { a, b, c, d }

extension _enumUIntExt on enumUInt {
  static enumUInt fromValue(int i) {
    return enumUInt.values[i];
  }

  int toValue() {
    return index;
  }
}

enumUInt enumUIntFromJson(dynamic json) => _enumUIntExt.fromValue(json as int);

dynamic enumUIntToJson(enumUInt item) => item.toValue();

int intFromJson(dynamic json) => json as int;

int intToJson(int item) => item;

// github.com/benoitkugler/gomacro/testutils/testsource.itfList
typedef itfList = List<itfType>;

/// github.com/benoitkugler/gomacro/testutils/testsource.itfType
abstract class itfType {}

itfType itfTypeFromJson(dynamic json_) {
  final json = json_ as JSON;
  final kind = json['Kind'] as String;
  final data = json['Data'];
  switch (kind) {
    case "concretType1":
      return concretType1FromJson(data);
    case "concretType2":
      return concretType2FromJson(data);
    default:
      throw ("unexpected type");
  }
}

JSON itfTypeToJson(itfType item) {
  if (item is concretType1) {
    return {'Kind': "concretType1", 'Data': concretType1ToJson(item)};
  } else if (item is concretType2) {
    return {'Kind': "concretType2", 'Data': concretType2ToJson(item)};
  } else {
    throw ("unexpected type");
  }
}

/// github.com/benoitkugler/gomacro/testutils/testsource.itfType2
abstract class itfType2 {}

itfType2 itfType2FromJson(dynamic json_) {
  final json = json_ as JSON;
  final kind = json['Kind'] as String;
  final data = json['Data'];
  switch (kind) {
    case "concretType1":
      return concretType1FromJson(data);
    default:
      throw ("unexpected type");
  }
}

JSON itfType2ToJson(itfType2 item) {
  if (item is concretType1) {
    return {'Kind': "concretType1", 'Data': concretType1ToJson(item)};
  } else {
    throw ("unexpected type");
  }
}

List<int> listIntFromJson(dynamic json) {
  if (json == null) {
    return [];
  }
  return (json as List<dynamic>).map(intFromJson).toList();
}

List<dynamic> listIntToJson(List<int> item) {
  return item.map(intToJson).toList();
}

List<itfType> listItfTypeFromJson(dynamic json) {
  if (json == null) {
    return [];
  }
  return (json as List<dynamic>).map(itfTypeFromJson).toList();
}

List<dynamic> listItfTypeToJson(List<itfType> item) {
  return item.map(itfTypeToJson).toList();
}

List<recursiveType> listRecursiveTypeFromJson(dynamic json) {
  if (json == null) {
    return [];
  }
  return (json as List<dynamic>).map(recursiveTypeFromJson).toList();
}

List<dynamic> listRecursiveTypeToJson(List<recursiveType> item) {
  return item.map(recursiveTypeToJson).toList();
}

// github.com/benoitkugler/gomacro/testutils/testsource.notAnEnum
typedef notAnEnum = String;

// github.com/benoitkugler/gomacro/testutils/testsource.recursiveType
class recursiveType {
  final List<recursiveType> children;

  const recursiveType(this.children);

  @override
  String toString() {
    return "recursiveType($children)";
  }
}

recursiveType recursiveTypeFromJson(dynamic json_) {
  final json = (json_ as JSON);
  return recursiveType(listRecursiveTypeFromJson(json['Children']));
}

JSON recursiveTypeToJson(recursiveType item) {
  return {"Children": listRecursiveTypeToJson(item.children)};
}

// github.com/benoitkugler/gomacro/testutils/testsource.structWithExternalRef
class structWithExternalRef {
  final CancelFunc field2;

  const structWithExternalRef(this.field2);

  @override
  String toString() {
    return "structWithExternalRef($field2)";
  }
}

structWithExternalRef structWithExternalRefFromJson(dynamic json_) {
  final json = (json_ as JSON);
  return structWithExternalRef(cancelFuncFromJson(json['Field2']));
}

JSON structWithExternalRefToJson(structWithExternalRef item) {
  return {"Field2": cancelFuncToJson(item.field2)};
}

// github.com/benoitkugler/gomacro/testutils/testsource.withEmbeded
class withEmbeded {
  final Map<int, int> dict;
  final int u;
  final DateTime time;
  final String b;
  final itfType value;
  final itfList l;
  final int a;
  final enumInt e;
  final enumUInt e2;
  final enumString e3;
  final MyDate date;
  final List<int> f;
  final StructWithComment imported;

  const withEmbeded(this.dict, this.u, this.time, this.b, this.value, this.l,
      this.a, this.e, this.e2, this.e3, this.date, this.f, this.imported);

  @override
  String toString() {
    return "withEmbeded($dict, $u, $time, $b, $value, $l, $a, $e, $e2, $e3, $date, $f, $imported)";
  }
}

withEmbeded withEmbededFromJson(dynamic json_) {
  final json = (json_ as JSON);
  return withEmbeded(
      dictInt_IntFromJson(json['Dict']),
      intFromJson(json['U']),
      dateTimeFromJson(json['Time']),
      stringFromJson(json['B']),
      itfTypeFromJson(json['Value']),
      listItfTypeFromJson(json['L']),
      intFromJson(json['A']),
      enumIntFromJson(json['E']),
      enumUIntFromJson(json['E2']),
      enumStringFromJson(json['E3']),
      dateTimeFromJson(json['Date']),
      listIntFromJson(json['F']),
      structWithCommentFromJson(json['Imported']));
}

JSON withEmbededToJson(withEmbeded item) {
  return {
    "Dict": dictInt_IntToJson(item.dict),
    "U": intToJson(item.u),
    "Time": dateTimeToJson(item.time),
    "B": stringToJson(item.b),
    "Value": itfTypeToJson(item.value),
    "L": listItfTypeToJson(item.l),
    "A": intToJson(item.a),
    "E": enumIntToJson(item.e),
    "E2": enumUIntToJson(item.e2),
    "E3": enumStringToJson(item.e3),
    "Date": dateTimeToJson(item.date),
    "F": listIntToJson(item.f),
    "Imported": structWithCommentToJson(item.imported)
  };
}
