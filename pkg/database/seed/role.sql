INSERT INTO roles(name, description)
SELECT * 
FROM (VALUES
  ('СУПЕРАДМИН', 'Главный администратор системы имеет полный доступ ко всем ресурсам')
) AS values_tobe_inserted(name, description)
WHERE NOT EXISTS (
  SELECT *
  FROM roles
  WHERE roles.name = values_tobe_inserted.name AND roles.description = values_tobe_inserted.description
);


