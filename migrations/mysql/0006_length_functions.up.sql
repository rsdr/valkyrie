CREATE OR REPLACE FUNCTION from_go_duration(IN d BIGINT) RETURNS INT return d DIV 1000000000;
CREATE OR REPLACE FUNCTION to_go_duration(IN s INT) RETURNS BIGINT return s * 1000000000;