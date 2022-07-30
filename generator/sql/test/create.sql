CREATE TABLE exercices (
    Id serial PRIMARY KEY,
    Title text NOT NULL,
    Description text NOT NULL,
    Parameters jsonb NOT NULL CONSTRAINT Parameters_gomacro_validate_json_map_boolean CHECK (gomacro_validate_json_map_boolean (Parameters)),
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
    History jsonb NOT NULL CONSTRAINT History_gomacro_validate_json_array_test_EnumUInt CHECK (gomacro_validate_json_array_test_EnumUInt (History))
);

CREATE TABLE questions (
    Id serial PRIMARY KEY,
    Page jsonb NOT NULL CONSTRAINT Page_gomacro_validate_json_test_ComplexStruct CHECK (gomacro_validate_json_test_ComplexStruct (Page)),
    Public boolean NOT NULL,
    IdTeacher integer NOT NULL,
    Description text NOT NULL,
    NeedExercice integer
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
    L integer
);

CREATE OR REPLACE FUNCTION gomacro_validate_json_array_5_number (data jsonb)
    RETURNS boolean
    AS $$
BEGIN
    IF jsonb_typeof(data) != 'array' THEN
        RETURN FALSE;
    END IF;
    RETURN (
        SELECT
            bool_and(gomacro_validate_json_number (value))
        FROM
            jsonb_array_elements(data))
        AND jsonb_array_length(data) = 5;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_array_number (data jsonb)
    RETURNS boolean
    AS $$
BEGIN
    IF jsonb_typeof(data) = 'null' THEN
        RETURN TRUE;
    END IF;
    IF jsonb_typeof(data) != 'array' THEN
        RETURN FALSE;
    END IF;
    IF jsonb_array_length(data) = 0 THEN
        RETURN TRUE;
    END IF;
    RETURN (
        SELECT
            bool_and(gomacro_validate_json_number (value))
        FROM
            jsonb_array_elements(data));
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_array_test_EnumUInt (data jsonb)
    RETURNS boolean
    AS $$
BEGIN
    IF jsonb_typeof(data) = 'null' THEN
        RETURN TRUE;
    END IF;
    IF jsonb_typeof(data) != 'array' THEN
        RETURN FALSE;
    END IF;
    IF jsonb_array_length(data) = 0 THEN
        RETURN TRUE;
    END IF;
    RETURN (
        SELECT
            bool_and(gomacro_validate_json_test_EnumUInt (value))
        FROM
            jsonb_array_elements(data));
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_array_test_ItfType (data jsonb)
    RETURNS boolean
    AS $$
BEGIN
    IF jsonb_typeof(data) = 'null' THEN
        RETURN TRUE;
    END IF;
    IF jsonb_typeof(data) != 'array' THEN
        RETURN FALSE;
    END IF;
    IF jsonb_array_length(data) = 0 THEN
        RETURN TRUE;
    END IF;
    RETURN (
        SELECT
            bool_and(gomacro_validate_json_test_ItfType (value))
        FROM
            jsonb_array_elements(data));
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_boolean (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean := jsonb_typeof(data) = 'boolean';
BEGIN
    IF NOT is_valid THEN
        RAISE WARNING '% is not a boolean', data;
    END IF;
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_map_boolean (data jsonb)
    RETURNS boolean
    AS $$
BEGIN
    IF jsonb_typeof(data) = 'null' THEN
        -- accept null value coming from nil maps
        RETURN TRUE;
    END IF;
    RETURN jsonb_typeof(data) = 'object'
        AND (
            SELECT
                bool_and(gomacro_validate_json_boolean (value))
            FROM
                jsonb_each(data));
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_map_number (data jsonb)
    RETURNS boolean
    AS $$
BEGIN
    IF jsonb_typeof(data) = 'null' THEN
        -- accept null value coming from nil maps
        RETURN TRUE;
    END IF;
    RETURN jsonb_typeof(data) = 'object'
        AND (
            SELECT
                bool_and(gomacro_validate_json_number (value))
            FROM
                jsonb_each(data));
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_number (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean := jsonb_typeof(data) = 'number';
BEGIN
    IF NOT is_valid THEN
        RAISE WARNING '% is not a number', data;
    END IF;
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_string (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean := jsonb_typeof(data) = 'string';
BEGIN
    IF NOT is_valid THEN
        RAISE WARNING '% is not a string', data;
    END IF;
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_subp_StructWithComment (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean;
BEGIN
    IF jsonb_typeof(data) != 'object' THEN
        RETURN FALSE;
    END IF;
    is_valid := (
        SELECT
            bool_and(key IN ('A'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_number (data -> 'A');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_test_ComplexStruct (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean;
BEGIN
    IF jsonb_typeof(data) != 'object' THEN
        RETURN FALSE;
    END IF;
    is_valid := (
        SELECT
            bool_and(key IN ('Dict', 'Time', 'B', 'Value', 'L', 'A', 'E', 'E2', 'Date', 'F', 'Imported'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_map_number (data -> 'Dict')
        AND gomacro_validate_json_string (data -> 'Time')
        AND gomacro_validate_json_string (data -> 'B')
        AND gomacro_validate_json_test_ItfType (data -> 'Value')
        AND gomacro_validate_json_test_ItfList (data -> 'L')
        AND gomacro_validate_json_number (data -> 'A')
        AND gomacro_validate_json_test_EnumInt (data -> 'E')
        AND gomacro_validate_json_test_EnumUInt (data -> 'E2')
        AND gomacro_validate_json_test_MyDate (data -> 'Date')
        AND gomacro_validate_json_array_5_number (data -> 'F')
        AND gomacro_validate_json_subp_StructWithComment (data -> 'Imported');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_test_ConcretType1 (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean;
BEGIN
    IF jsonb_typeof(data) != 'object' THEN
        RETURN FALSE;
    END IF;
    is_valid := (
        SELECT
            bool_and(key IN ('List2', 'V'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_array_number (data -> 'List2')
        AND gomacro_validate_json_number (data -> 'V');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_test_ConcretType2 (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean;
BEGIN
    IF jsonb_typeof(data) != 'object' THEN
        RETURN FALSE;
    END IF;
    is_valid := (
        SELECT
            bool_and(key IN ('D'))
        FROM
            jsonb_each(data))
        AND gomacro_validate_json_number (data -> 'D');
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_test_EnumInt (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean := jsonb_typeof(data) = 'number'
    AND data::int IN (0, 1, 2, 4);
BEGIN
    IF NOT is_valid THEN
        RAISE WARNING '% is not a test_EnumInt', data;
    END IF;
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_test_EnumUInt (data jsonb)
    RETURNS boolean
    AS $$
DECLARE
    is_valid boolean := jsonb_typeof(data) = 'number'
    AND data::int IN (0, 1, 2, 3, 4);
BEGIN
    IF NOT is_valid THEN
        RAISE WARNING '% is not a test_EnumUInt', data;
    END IF;
    RETURN is_valid;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

CREATE OR REPLACE FUNCTION gomacro_validate_json_test_ItfType (data jsonb)
    RETURNS boolean
    AS $$
BEGIN
    IF jsonb_typeof(data) != 'object' OR jsonb_typeof(data -> 'Kind') != 'string' OR jsonb_typeof(data -> 'Data') = 'null' THEN
        RETURN FALSE;
    END IF;
    CASE WHEN data ->> 'Kind' = 'ConcretType1' THEN
        RETURN gomacro_validate_json_test_ConcretType1 (data -> 'Data');
    WHEN data ->> 'Kind' = 'ConcretType2' THEN
        RETURN gomacro_validate_json_test_ConcretType2 (data -> 'Data');
    ELSE
        RETURN FALSE;
    END CASE;
END;
$$
LANGUAGE 'plpgsql'
IMMUTABLE;

