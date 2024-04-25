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

INSERT INTO permissions(role_id, resource_id, r, w, u, d)
SELECT *
FROM (VALUES
  (1, 1, true, true, true, true),
  (1, 2, true, true, true, true),
  (1, 3, true, true, true, true),
  (1, 4, true, true, true, true),
  (1, 5, true, true, true, true),
  (1, 6, true, true, true, true),
  (1, 7, true, true, true, true),
  (1, 8, true, true, true, true),
  (1, 9, true, true, true, true),
  (1, 10, true, true, true, true),
  (1, 11, true, true, true, true),
  (1, 12, true, true, true, true),
  (1, 13, true, true, true, true),
  (1, 14, true, true, true, true),
  (1, 15, true, true, true, true),
  (1, 16, true, true, true, true),
  (1, 17, true, true, true, true),
  (1, 18, true, true, true, true),
  (1, 19, true, true, true, true),
  (1, 20, true, true, true, true),
  (1, 21, true, true, true, true),
  (1, 22, true, true, true, true),
  (1, 23, true, true, true, true),
  (1, 24, true, true, true, true)
) AS values_tobe_inserted(role_id, resource_id, r, w, u, d)
WHERE NOT EXISTS (
  SELECT *
  FROM permissions
  WHERE  
    permissions.role_id = values_tobe_inserted.role_id
    AND permissions.resource_id = values_tobe_inserted.resource_id
    AND permissions.r = values_tobe_inserted.r
    AND permissions.w = values_tobe_inserted.w
    AND permissions.u = values_tobe_inserted.u
    AND permissions.u = values_tobe_inserted.d
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


