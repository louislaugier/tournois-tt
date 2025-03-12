-- Fix Windows-1252 encoding issues in the database
-- Migration to convert incorrectly encoded characters to proper UTF-8

-- Function to fix Windows-1252 encoding
CREATE OR REPLACE FUNCTION fix_windows1252_encoding(text_to_fix TEXT)
RETURNS TEXT AS $$
DECLARE
    result TEXT;
BEGIN
    result := text_to_fix;
    
    -- Handle two-byte sequences first (longer patterns take precedence)
    -- Each pattern handles both lowercase and uppercase based on word position
    
    -- é/É
    result := regexp_replace(
        result,
        '(^|\\s)\\xc3\\xa9',
        E'\\1\\x{00c9}',
        'g'
    );
    result := regexp_replace(
        result,
        '\\xc3\\xa9',
        E'\\x{00e9}',
        'g'
    );
    
    -- í/Í
    result := regexp_replace(
        result,
        '(^|\\s)\\xc3\\xad',
        E'\\1\\x{00cd}',
        'g'
    );
    result := regexp_replace(
        result,
        '\\xc3\\xad',
        E'\\x{00ed}',
        'g'
    );
    
    -- á/Á
    result := regexp_replace(
        result,
        '(^|\\s)\\xc3\\xa1',
        E'\\1\\x{00c1}',
        'g'
    );
    result := regexp_replace(
        result,
        '\\xc3\\xa1',
        E'\\x{00e1}',
        'g'
    );
    
    -- ó/Ó
    result := regexp_replace(
        result,
        '(^|\\s)\\xc3\\xb3',
        E'\\1\\x{00d3}',
        'g'
    );
    result := regexp_replace(
        result,
        '\\xc3\\xb3',
        E'\\x{00f3}',
        'g'
    );
    
    -- ñ/Ñ
    result := regexp_replace(
        result,
        '(^|\\s)\\xc3\\xb1',
        E'\\1\\x{00d1}',
        'g'
    );
    result := regexp_replace(
        result,
        '\\xc3\\xb1',
        E'\\x{00f1}',
        'g'
    );
    
    -- ú/Ú
    result := regexp_replace(
        result,
        '(^|\\s)\\xc3\\xba',
        E'\\1\\x{00da}',
        'g'
    );
    result := regexp_replace(
        result,
        '\\xc3\\xba',
        E'\\x{00fa}',
        'g'
    );
    
    -- Handle single character followed by a letter
    -- ía/Ía
    result := regexp_replace(
        result,
        '(^|\\s)\\xc3a',
        E'\\1\\x{00cd}a',
        'g'
    );
    result := regexp_replace(
        result,
        '\\xc3a',
        E'\\x{00ed}a',
        'g'
    );
    
    -- é/É
    result := regexp_replace(
        result,
        '(^|\\s)\\xc3e',
        E'\\1\\x{00c9}',
        'g'
    );
    result := regexp_replace(
        result,
        '\\xc3e',
        E'\\x{00e9}',
        'g'
    );
    
    -- í/Í
    result := regexp_replace(
        result,
        '(^|\\s)\\xc3i',
        E'\\1\\x{00cd}',
        'g'
    );
    result := regexp_replace(
        result,
        '\\xc3i',
        E'\\x{00ed}',
        'g'
    );
    
    -- ó/Ó
    result := regexp_replace(
        result,
        '(^|\\s)\\xc3o',
        E'\\1\\x{00d3}',
        'g'
    );
    result := regexp_replace(
        result,
        '\\xc3o',
        E'\\x{00f3}',
        'g'
    );
    
    -- ú/Ú
    result := regexp_replace(
        result,
        '(^|\\s)\\xc3u',
        E'\\1\\x{00da}',
        'g'
    );
    result := regexp_replace(
        result,
        '\\xc3u',
        E'\\x{00fa}',
        'g'
    );
    
    -- ñ/Ñ
    result := regexp_replace(
        result,
        '(^|\\s)\\xc3n',
        E'\\1\\x{00d1}',
        'g'
    );
    result := regexp_replace(
        result,
        '\\xc3n',
        E'\\x{00f1}',
        'g'
    );
    
    -- Handle lone Ã (default to í/Í)
    result := regexp_replace(
        result,
        '(^|\\s)\\xc3(?![\\xa9\\xad\\xa1\\xb3\\xb1\\xba]|[aeioun])',
        E'\\1\\x{00cd}',
        'g'
    );
    result := regexp_replace(
        result,
        '\\xc3(?![\\xa9\\xad\\xa1\\xb3\\xb1\\xba]|[aeioun])',
        E'\\x{00ed}',
        'g'
    );
    
    RETURN result;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Example usage for a table 'users' with column 'name':
-- UPDATE users SET name = fix_windows1252_encoding(name)
-- WHERE name ~ E'\\xc3([\\xa9\\xad\\xa1\\xb3\\xb1\\xba]|[aeioun]|$)';

-- NOTE: Replace table_name and column_name with your actual table and column names
-- DO $$
-- DECLARE
--     r RECORD;
-- BEGIN
--     FOR r IN SELECT table_name, column_name 
--              FROM information_schema.columns 
--              WHERE data_type IN ('character varying', 'text')
--              AND table_schema = 'public'
--     LOOP
--         EXECUTE format('UPDATE %I SET %I = fix_windows1252_encoding(%I) WHERE %I ~ E''\\xc3([\\xa9\\xad\\xa1\\xb3\\xb1\\xba]|[aeioun]|$)''',
--             r.table_name, r.column_name, r.column_name, r.column_name);
--     END LOOP;
-- END $$;

-- To revert (if needed):
-- DROP FUNCTION IF EXISTS fix_windows1252_encoding(TEXT); 