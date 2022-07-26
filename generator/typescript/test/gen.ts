import { CancelFunc } from "./extern";
// github.com/benoitkugler/gomacro/testutils/testsource.MyDate
export type MyDate = Date_;
// github.com/benoitkugler/gomacro/testutils/testsource/subpackage.StructWithComment
export interface StructWithComment {
  A: number;
}

class DateTag {
  private _: "D" = "D";
}

// AAAA-MM-YY date format
export type Date_ = string & DateTag;

class TimeTag {
  private _: "T" = "T";
}

// ISO date-time string
export type Time = string & TimeTag;

// github.com/benoitkugler/gomacro/testutils/testsource.basic1
export type basic1 = number;
// github.com/benoitkugler/gomacro/testutils/testsource.basic2
export type basic2 = boolean;
// github.com/benoitkugler/gomacro/testutils/testsource.basic3
export type basic3 = number;
// github.com/benoitkugler/gomacro/testutils/testsource.basic4
export type basic4 = string;
// github.com/benoitkugler/gomacro/testutils/testsource.complexStruct
export interface complexStruct {
  Dict: { [key: number]: number } | null;
  U: number;
  Time: Time;
  B: string;
  Value: itfType;
  L: itfList;
  A: number;
  E: enumInt;
  E2: enumUInt;
  E3: enumString;
  Date: MyDate;
  F: number[];
  Imported: StructWithComment;
}
// github.com/benoitkugler/gomacro/testutils/testsource.concretType1
export interface concretType1 {
  List2: number[] | null;
  V: number;
}
// github.com/benoitkugler/gomacro/testutils/testsource.concretType2
export interface concretType2 {
  D: number;
}
// github.com/benoitkugler/gomacro/testutils/testsource.enumInt
export enum enumInt {
  Ai = 0,
  Bi = 1,
  Ci = 2,
  Di = 4,
}

export const enumIntLabels: { [key in enumInt]: string } = {
  [enumInt.Ai]: "sdsd",
  [enumInt.Bi]: "sdsdB",
  [enumInt.Ci]: "sdsdC",
  [enumInt.Di]: "sdsdD",
};

// github.com/benoitkugler/gomacro/testutils/testsource.enumString
export enum enumString {
  SA = "va",
  SB = "vb",
  SC = "vc",
  SD = "vd",
}

export const enumStringLabels: { [key in enumString]: string } = {
  [enumString.SA]: "sddA",
  [enumString.SB]: "sddB",
  [enumString.SC]: "sddC",
  [enumString.SD]: "sddD",
};

// github.com/benoitkugler/gomacro/testutils/testsource.enumUInt
export enum enumUInt {
  A = 0,
  B = 1,
  C = 2,
  D = 3,
  e = 4,
}

export const enumUIntLabels: { [key in enumUInt]: string } = {
  [enumUInt.A]: "sdsd",
  [enumUInt.B]: "sdsdB",
  [enumUInt.C]: "sdsdC",
  [enumUInt.D]: "sdsdD",
  [enumUInt.e]: "not added",
};

// github.com/benoitkugler/gomacro/testutils/testsource.itfList
export type itfList = itfType[] | null;

export enum itfTypeKind {
  concretType1 = "concretType1",
  concretType2 = "concretType2",
}

// github.com/benoitkugler/gomacro/testutils/testsource.itfType
export interface itfType {
  Kind: itfTypeKind;
  Data: concretType1 | concretType2;
}

export enum itfType2Kind {
  concretType1 = "concretType1",
}

// github.com/benoitkugler/gomacro/testutils/testsource.itfType2
export interface itfType2 {
  Kind: itfType2Kind;
  Data: concretType1;
}
// github.com/benoitkugler/gomacro/testutils/testsource.notAnEnum
export type notAnEnum = string;
// github.com/benoitkugler/gomacro/testutils/testsource.recursiveType
export interface recursiveType {
  Children: recursiveType[] | null;
}
// github.com/benoitkugler/gomacro/testutils/testsource.structWithExternalRef
export interface structWithExternalRef {
  Field2: CancelFunc;
}
// github.com/benoitkugler/gomacro/testutils/testsource.withEmbeded
export interface withEmbeded {
  Dict: { [key: number]: number } | null;
  U: number;
  Time: Time;
  B: string;
  Value: itfType;
  L: itfList;
  A: number;
  E: enumInt;
  E2: enumUInt;
  E3: enumString;
  Date: MyDate;
  F: number[];
  Imported: StructWithComment;
}
