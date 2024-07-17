INSERT INTO roles(name, description)
SELECT * 
FROM (VALUES
  ('Суперадмин', 'Главный администратор системы имеет полный доступ ко всем ресурсам')
) AS values_tobe_inserted(name, description)
WHERE NOT EXISTS (
  SELECT *
  FROM roles
  WHERE roles.name = values_tobe_inserted.name AND roles.description = values_tobe_inserted.description
);

INSERT INTO workers(name, job_title, mobile_number)
SELECT * 
FROM (VALUES
  ('Суперадмин', 'Главный администратор системы', '+9929999999')
) AS values_tobe_inserted(name, job_title, mobile_number)
WHERE NOT EXISTS (
  SELECT *
  FROM workers
  WHERE 
    workers.name = values_tobe_inserted.name 
    AND workers.job_title = values_tobe_inserted.job_title
    AND workers.mobile_number = values_tobe_inserted.mobile_number
);

INSERT INTO users(worker_id, role_id, username, password)
SELECT * 
FROM (VALUES
  (1, 1, 'superadmin','$2a$10$2kZzCVY9TaX4Uy5NEQVlP.RNhsPOy.yRzvU08YYWbbs2Sk0D0O5Sy')
) AS values_tobe_inserted(worker_id, role_id, username, password)
WHERE NOT EXISTS (
  SELECT *
  FROM users
  WHERE 
    users.worker_id = values_tobe_inserted.worker_id 
    AND users.role_id = values_tobe_inserted.role_id
    AND users.username = values_tobe_inserted.username
    AND users.password = values_tobe_inserted.password
);

INSERT INTO user_in_projects(project_id, user_id)
SELECT * 
FROM (VALUES
  (1, 1)
) AS values_tobe_inserted(project_id, user_id)
WHERE NOT EXISTS (
  SELECT *
  FROM user_in_projects
  WHERE 
    user_in_projects.project_id = values_tobe_inserted.project_id 
    AND user_in_projects.user_id = values_tobe_inserted.user_id
);


