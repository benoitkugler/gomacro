// Code generated by gomacro/generator/dart. DO NOT EDIT

import 'predefined.dart';
import 'testsource_subpackage.dart';

// github.com/benoitkugler/gomacro/testutils/testsource.Basic1
typedef Basic1 = int;

// github.com/benoitkugler/gomacro/testutils/testsource.Basic2
typedef Basic2 = bool;

// github.com/benoitkugler/gomacro/testutils/testsource.Basic3
typedef Basic3 = double;

// github.com/benoitkugler/gomacro/testutils/testsource.Basic4
typedef Basic4 = String;

// github.com/benoitkugler/gomacro/testutils/testsource.ComplexStruct
class ComplexStruct {
  final Map<int, int> with_tag;
  final DateTime time;
  final String b;
  final ItfType value;
  final ItfList l;
  final int a;
  final EnumInt e;
  final EnumUInt e2;
  final MyDate date;
  final List<int> f;
  final StructWithComment imported;
  final Map<EnumInt, bool> enumMap;

  const ComplexStruct(this.with_tag, this.time, this.b, this.value, this.l,
      this.a, this.e, this.e2, this.date, this.f, this.imported, this.enumMap);

  @override
  String toString() {
    return "ComplexStruct($with_tag, $time, $b, $value, $l, $a, $e, $e2, $date, $f, $imported, $enumMap)";
  }
}

ComplexStruct complexStructFromJson(dynamic json_) {
  final json = (json_ as Map<String, dynamic>);
  return ComplexStruct(
      dictIntToIntFromJson(json['with_tag']),
      dateTimeFromJson(json['Time']),
      stringFromJson(json['B']),
      itfTypeFromJson(json['Value']),
      itfListFromJson(json['L']),
      intFromJson(json['A']),
      enumIntFromJson(json['E']),
      enumUIntFromJson(json['E2']),
      dateTimeFromJson(json['Date']),
      listIntFromJson(json['F']),
      structWithCommentFromJson(json['Imported']),
      dictEnumIntToBoolFromJson(json['EnumMap']));
}

Map<String, dynamic> complexStructToJson(ComplexStruct item) {
  return {
    "with_tag": dictIntToIntToJson(item.with_tag),
    "Time": dateTimeToJson(item.time),
    "B": stringToJson(item.b),
    "Value": itfTypeToJson(item.value),
    "L": itfListToJson(item.l),
    "A": intToJson(item.a),
    "E": enumIntToJson(item.e),
    "E2": enumUIntToJson(item.e2),
    "Date": dateTimeToJson(item.date),
    "F": listIntToJson(item.f),
    "Imported": structWithCommentToJson(item.imported),
    "EnumMap": dictEnumIntToBoolToJson(item.enumMap)
  };
}

// github.com/benoitkugler/gomacro/testutils/testsource.ConcretType1
class ConcretType1 implements ItfType, ItfType2 {
  final List<int> list2;
  final int v;

  const ConcretType1(this.list2, this.v);

  @override
  String toString() {
    return "ConcretType1($list2, $v)";
  }
}

ConcretType1 concretType1FromJson(dynamic json_) {
  final json = (json_ as Map<String, dynamic>);
  return ConcretType1(listIntFromJson(json['List2']), intFromJson(json['V']));
}

Map<String, dynamic> concretType1ToJson(ConcretType1 item) {
  return {"List2": listIntToJson(item.list2), "V": intToJson(item.v)};
}

// github.com/benoitkugler/gomacro/testutils/testsource.ConcretType2
class ConcretType2 implements ItfType {
  final double d;

  const ConcretType2(this.d);

  @override
  String toString() {
    return "ConcretType2($d)";
  }
}

ConcretType2 concretType2FromJson(dynamic json_) {
  final json = (json_ as Map<String, dynamic>);
  return ConcretType2(doubleFromJson(json['D']));
}

Map<String, dynamic> concretType2ToJson(ConcretType2 item) {
  return {"D": doubleToJson(item.d)};
}

// github.com/benoitkugler/gomacro/testutils/testsource.EnumInt
enum EnumInt { ai, bi, ci, di }

extension _EnumIntExt on EnumInt {
  static const _values = [0, 1, 2, 4];
  static EnumInt fromValue(int s) {
    return EnumInt.values[_values.indexOf(s)];
  }

  int toValue() {
    return _values[index];
  }
}

String enumIntLabel(EnumInt v) {
  switch (v) {
    case EnumInt.ai:
      return "sdsd";
    case EnumInt.bi:
      return "sdsdB";
    case EnumInt.ci:
      return "sdsdC";
    case EnumInt.di:
      return "sdsdD";
  }
}

EnumInt enumIntFromJson(dynamic json) => _EnumIntExt.fromValue(json as int);

dynamic enumIntToJson(EnumInt item) => item.toValue();

// github.com/benoitkugler/gomacro/testutils/testsource.EnumUInt
enum EnumUInt { a, b, c, d }

extension _EnumUIntExt on EnumUInt {
  static EnumUInt fromValue(int i) {
    return EnumUInt.values[i];
  }

  int toValue() {
    return index;
  }
}

String enumUIntLabel(EnumUInt v) {
  switch (v) {
    case EnumUInt.a:
      return "sdsd";
    case EnumUInt.b:
      return "sdsdB";
    case EnumUInt.c:
      return "sdsdC";
    case EnumUInt.d:
      return "sdsdD";
  }
}

EnumUInt enumUIntFromJson(dynamic json) => _EnumUIntExt.fromValue(json as int);

dynamic enumUIntToJson(EnumUInt item) => item.toValue();

// github.com/benoitkugler/gomacro/testutils/testsource.ItfList
typedef ItfList = List<ItfType>;

ItfList itfListFromJson(dynamic json) {
  return listItfTypeFromJson(json);
}

dynamic itfListToJson(ItfList item) {
  return listItfTypeToJson(item);
}

/// github.com/benoitkugler/gomacro/testutils/testsource.ItfType
abstract class ItfType {}

ItfType itfTypeFromJson(dynamic json_) {
  final json = json_ as Map<String, dynamic>;
  final kind = json['Kind'] as String;
  final data = json['Data'];
  switch (kind) {
    case "ConcretType1":
      return concretType1FromJson(data);
    case "ConcretType2":
      return concretType2FromJson(data);
    default:
      throw ("unexpected type");
  }
}

Map<String, dynamic> itfTypeToJson(ItfType item) {
  if (item is ConcretType1) {
    return {'Kind': "ConcretType1", 'Data': concretType1ToJson(item)};
  } else if (item is ConcretType2) {
    return {'Kind': "ConcretType2", 'Data': concretType2ToJson(item)};
  } else {
    throw ("unexpected type");
  }
}

/// github.com/benoitkugler/gomacro/testutils/testsource.ItfType2
abstract class ItfType2 {}

ItfType2 itfType2FromJson(dynamic json_) {
  final json = json_ as Map<String, dynamic>;
  final kind = json['Kind'] as String;
  final data = json['Data'];
  switch (kind) {
    case "ConcretType1":
      return concretType1FromJson(data);
    default:
      throw ("unexpected type");
  }
}

Map<String, dynamic> itfType2ToJson(ItfType2 item) {
  if (item is ConcretType1) {
    return {'Kind': "ConcretType1", 'Data': concretType1ToJson(item)};
  } else {
    throw ("unexpected type");
  }
}

// github.com/benoitkugler/gomacro/testutils/testsource.MyDate
typedef MyDate = DateTime;

// github.com/benoitkugler/gomacro/testutils/testsource.RecursiveType
class RecursiveType {
  final List<RecursiveType> children;

  const RecursiveType(this.children);

  @override
  String toString() {
    return "RecursiveType($children)";
  }
}

RecursiveType recursiveTypeFromJson(dynamic json_) {
  final json = (json_ as Map<String, dynamic>);
  return RecursiveType(listRecursiveTypeFromJson(json['Children']));
}

Map<String, dynamic> recursiveTypeToJson(RecursiveType item) {
  return {"Children": listRecursiveTypeToJson(item.children)};
}

// github.com/benoitkugler/gomacro/testutils/testsource.StructWithExternalRef
class StructWithExternalRef {
  final NamedSlice field1;
  final NamedSlice field2;
  final int field3;

  const StructWithExternalRef(this.field1, this.field2, this.field3);

  @override
  String toString() {
    return "StructWithExternalRef($field1, $field2, $field3)";
  }
}

StructWithExternalRef structWithExternalRefFromJson(dynamic json_) {
  final json = (json_ as Map<String, dynamic>);
  return StructWithExternalRef(namedSliceFromJson(json['Field1']),
      namedSliceFromJson(json['Field2']), intFromJson(json['Field3']));
}

Map<String, dynamic> structWithExternalRefToJson(StructWithExternalRef item) {
  return {
    "Field1": namedSliceToJson(item.field1),
    "Field2": namedSliceToJson(item.field2),
    "Field3": intToJson(item.field3)
  };
}

// github.com/benoitkugler/gomacro/testutils/testsource.WithOpaque
class WithOpaque {
  final dynamic f1;
  final dynamic f2;
  final StructWithExternalRef f3;

  const WithOpaque(this.f1, this.f2, this.f3);

  @override
  String toString() {
    return "WithOpaque($f1, $f2, $f3)";
  }
}

WithOpaque withOpaqueFromJson(dynamic json_) {
  final json = (json_ as Map<String, dynamic>);
  return WithOpaque(
      json['F1'], json['F2'], structWithExternalRefFromJson(json['F3']));
}

Map<String, dynamic> withOpaqueToJson(WithOpaque item) {
  return {
    "F1": item.f1,
    "F2": item.f2,
    "F3": structWithExternalRefToJson(item.f3)
  };
}

Map<EnumInt, bool> dictEnumIntToBoolFromJson(dynamic json) {
  if (json == null) {
    return {};
  }
  return (json as Map<String, dynamic>)
      .map((k, v) => MapEntry(k as EnumInt, boolFromJson(v)));
}

Map<String, dynamic> dictEnumIntToBoolToJson(Map<EnumInt, bool> item) {
  return item
      .map((k, v) => MapEntry(enumIntToJson(k).toString(), boolToJson(v)));
}

Map<int, int> dictIntToIntFromJson(dynamic json) {
  if (json == null) {
    return {};
  }
  return (json as Map<String, dynamic>)
      .map((k, v) => MapEntry(int.parse(k), intFromJson(v)));
}

Map<String, dynamic> dictIntToIntToJson(Map<int, int> item) {
  return item.map((k, v) => MapEntry(intToJson(k).toString(), intToJson(v)));
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

List<ItfType> listItfTypeFromJson(dynamic json) {
  if (json == null) {
    return [];
  }
  return (json as List<dynamic>).map(itfTypeFromJson).toList();
}

List<dynamic> listItfTypeToJson(List<ItfType> item) {
  return item.map(itfTypeToJson).toList();
}

List<RecursiveType> listRecursiveTypeFromJson(dynamic json) {
  if (json == null) {
    return [];
  }
  return (json as List<dynamic>).map(recursiveTypeFromJson).toList();
}

List<dynamic> listRecursiveTypeToJson(List<RecursiveType> item) {
  return item.map(recursiveTypeToJson).toList();
}
