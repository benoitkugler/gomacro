import 'dart:convert';

import 'gen.dart';

main(List<String> args) {
  final m = complexStruct(
    {3: 4},
    5,
    DateTime.now(),
    "dsds",
    concretType1([1, 2, -1], 4),
    [concretType2(0.4), concretType2(0.8)],
    789,
    enumInt.bi,
    enumUInt.c,
    enumString.sB,
    DateTime.now(),
    [456, 456, 456],
    StructWithComment(5),
  );
  final json = complexStructToJson(m);
  final s = jsonEncode(json);
  print(s);

  final decoded = jsonDecode(s);
  final s2 = jsonEncode(complexStructToJson(complexStructFromJson(decoded)));

  if (s != s2) {
    throw ("inconstistent roundtrip");
  }
  print("OK");
}
