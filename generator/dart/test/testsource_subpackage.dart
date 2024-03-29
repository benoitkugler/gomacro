// Code generated by gomacro/generator/dart. DO NOT EDIT

import 'predefined.dart';

// github.com/benoitkugler/gomacro/testutils/testsource/subpackage.Enum
enum Enum { a, b, c }

extension _EnumExt on Enum {
  static Enum fromValue(int i) {
    return Enum.values[i];
  }

  int toValue() {
    return index;
  }
}

String enumLabel(Enum v) {
  switch (v) {
    case Enum.a:
      return "";
    case Enum.b:
      return "";
    case Enum.c:
      return "";
  }
}

Enum enumFromJson(dynamic json) => _EnumExt.fromValue(json as int);

dynamic enumToJson(Enum item) => item.toValue();

// github.com/benoitkugler/gomacro/testutils/testsource/subpackage.NamedSlice
typedef NamedSlice = List<Enum>;

NamedSlice namedSliceFromJson(dynamic json) {
  return listEnumFromJson(json);
}

dynamic namedSliceToJson(NamedSlice item) {
  return listEnumToJson(item);
}

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
  final json = (json_ as Map<String, dynamic>);
  return StructWithComment(intFromJson(json['A']));
}

Map<String, dynamic> structWithCommentToJson(StructWithComment item) {
  return {"A": intToJson(item.a)};
}

List<Enum> listEnumFromJson(dynamic json) {
  if (json == null) {
    return [];
  }
  return (json as List<dynamic>).map(enumFromJson).toList();
}

List<dynamic> listEnumToJson(List<Enum> item) {
  return item.map(enumToJson).toList();
}
