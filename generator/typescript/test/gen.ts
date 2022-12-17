class DateTag {
  private _ = "D" as const;
}

// AAAA-MM-YY date format
export type Date_ = string & DateTag;

class TimeTag {
  private _ = "T" as const;
}

// ISO date-time string
export type Time = string & TimeTag;

// github.com/benoitkugler/gomacro/testutils/testsource.Basic1
export type Basic1 = number;
// github.com/benoitkugler/gomacro/testutils/testsource.Basic2
export type Basic2 = boolean;
// github.com/benoitkugler/gomacro/testutils/testsource.Basic3
export type Basic3 = number;
// github.com/benoitkugler/gomacro/testutils/testsource.Basic4
export type Basic4 = string;
// github.com/benoitkugler/gomacro/testutils/testsource.ComplexStruct
export interface ComplexStruct {
  with_tag: { [key: number]: number } | null;
  Time: Time;
  B: string;
  Value: ItfType;
  L: ItfList;
  A: number;
  E: EnumInt;
  E2: EnumUInt;
  Date: MyDate;
  F: number[];
  Imported: StructWithComment;
}
// github.com/benoitkugler/gomacro/testutils/testsource.ConcretType1
export interface ConcretType1 {
  List2: number[] | null;
  V: number;
}
// github.com/benoitkugler/gomacro/testutils/testsource.ConcretType2
export interface ConcretType2 {
  D: number;
}
// github.com/benoitkugler/gomacro/testutils/testsource.EnumInt
export enum EnumInt {
  Ai = 0,
  Bi = 1,
  Ci = 2,
  Di = 4,
}

export const EnumIntLabels: { [key in EnumInt]: string } = {
  [EnumInt.Ai]: "sdsd",
  [EnumInt.Bi]: "sdsdB",
  [EnumInt.Ci]: "sdsdC",
  [EnumInt.Di]: "sdsdD",
};

// github.com/benoitkugler/gomacro/testutils/testsource.EnumUInt
export enum EnumUInt {
  A = 0,
  B = 1,
  C = 2,
  D = 3,
  e = 4,
}

export const EnumUIntLabels: { [key in EnumUInt]: string } = {
  [EnumUInt.A]: "sdsd",
  [EnumUInt.B]: "sdsdB",
  [EnumUInt.C]: "sdsdC",
  [EnumUInt.D]: "sdsdD",
  [EnumUInt.e]: "not added",
};

// github.com/benoitkugler/gomacro/testutils/testsource.ItfList
export type ItfList = ItfType[] | null;

export enum ItfTypeKind {
  ConcretType1 = "ConcretType1",
  ConcretType2 = "ConcretType2",
}

// github.com/benoitkugler/gomacro/testutils/testsource.ItfType
export interface ItfType {
  Kind: ItfTypeKind;
  Data: ConcretType1 | ConcretType2;
}

export enum ItfType2Kind {
  ConcretType1 = "ConcretType1",
}

// github.com/benoitkugler/gomacro/testutils/testsource.ItfType2
export interface ItfType2 {
  Kind: ItfType2Kind;
  Data: ConcretType1;
}
// github.com/benoitkugler/gomacro/testutils/testsource.MyDate
export type MyDate = Date_;
// github.com/benoitkugler/gomacro/testutils/testsource.RecursiveType
export interface RecursiveType {
  Children: RecursiveType[] | null;
}
// github.com/benoitkugler/gomacro/testutils/testsource.StructWithExternalRef
export interface StructWithExternalRef {
  Field1: NamedSlice;
  Field2: NamedSlice;
  Field3: number;
}
// github.com/benoitkugler/gomacro/testutils/testsource.WithOpaque
export interface WithOpaque {
  F1: StructWithExternalRef;
  F2: unknown;
  F3: unknown;
}
// github.com/benoitkugler/gomacro/testutils/testsource/subpackage.Enum
export enum Enum {
  A = 0,
  B = 1,
  C = 2,
}

export const EnumLabels: { [key in Enum]: string } = {
  [Enum.A]: "",
  [Enum.B]: "",
  [Enum.C]: "",
};

// github.com/benoitkugler/gomacro/testutils/testsource/subpackage.NamedSlice
export type NamedSlice = Enum[] | null;
// github.com/benoitkugler/gomacro/testutils/testsource/subpackage.StructWithComment
export interface StructWithComment {
  A: number;
}
