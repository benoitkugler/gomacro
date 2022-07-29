CREATE TABLE exercices (
    Id serial PRIMARY KEY,
    Title text NOT NULL,
    Description text NOT NULL,
    Parameters jsonb NOT NULL CONSTRAINT Parameters_ CHECK ((Parameters)),
    Flow integer CHECK (Flow IN (0, 1, 2, 4)) NOT NULL,
    IdTeacher integer NOT NULL,
    Public boolean NOT NULL
);

CREATE TABLE exercice_questions (
    IdExercice integer NOT NULL,
    IdQuestion integer NOT NULL,
    Bareme integer NOT NULL,
    Index integer NOT NULL
);

CREATE TABLE links ();

CREATE TABLE progressions (
    Id serial PRIMARY KEY,
    IdExercice integer NOT NULL
);

CREATE TABLE progression_questions (
    IdProgression integer NOT NULL,
    IdExercice integer NOT NULL,
    Index integer NOT NULL,
    History jsonb NOT NULL CONSTRAINT History_ CHECK ((History))
);

CREATE TABLE questions (
    Id serial PRIMARY KEY,
    Page jsonb NOT NULL CONSTRAINT Page_ CHECK ((Page)),
    Public boolean NOT NULL,
    IdTeacher integer NOT NULL,
    Description text NOT NULL,
    NeedExercice jsonb NOT NULL CONSTRAINT NeedExercice_ CHECK ((NeedExercice))
);

CREATE TABLE question_tags (
    Tag text NOT NULL,
    IdQuestion integer NOT NULL
);

CREATE TABLE repass ( Order text NOT NULL,
    Id serial PRIMARY KEY
);

CREATE TABLE table1s (
    Id serial PRIMARY KEY,
    Ex1 integer NOT NULL,
    Ex2 integer NOT NULL,
    L jsonb NOT NULL CONSTRAINT L_ CHECK ((L))
);

