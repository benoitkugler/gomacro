import 'dart:convert';

import 'testsource.dart';
import 'testsource_subpackage.dart';

main(List<String> args) {
  final m = ComplexStruct(
      {3: 4},
      DateTime.now(),
      "dsds",
      ConcretType1([1, 2, -1], 4),
      [ConcretType2(0.4), ConcretType2(0.8)],
      789,
      EnumInt.bi,
      EnumUInt.c,
      DateTime.now(),
      [
        [true],
        [false, true],
        []
      ],
      StructWithComment(5),
      {});
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
