INSERT INTO permissions(role_id, resource_name, resource_url, r, w, u, d)
SELECT *
FROM (VALUES
  (1, 'Накладная приход', '/input', true, true, true, true),
  (1, 'Накладная отпуск', '/output', true, true, true, true),
  (1, 'Накладная возврат', '/return', true, true, true, true),
  (1, 'Накладная списание', '/write-off', true, true, true, true),
  (1, 'Накладная объект', '/object', true, true, true, true),
  (1, 'Корректировка оператора', '/correction', true, true, true, true),
  (1, 'администрирование пользователями', '/user', true, true, true, true),
  (1, 'администрирование действия пользователей', '/user-action', true, true, true, true),
  (1, 'администрирование доступами пользователей в проекты', '/user-in-projects', true, true, true, true),
  (1, 'Администрирование ролями', '/role', true, true, true, true),
  (1, 'Администрирование доступами', '/permission', true, true, true, true),
  (1, 'Справочник материалов', '/material', true, true, true, true),
  (1, 'Ценники материалов', '/material-cost', true, true, true, true),
  (1, 'Справочник проектов', '/project', true, true, true, true),
  (1, 'Справочник районов', '/district', true, true, true, true),
  (1, 'Справочник сервисов', '/operation', true, true, true, true),
  (1, 'Справочник объектов', '/object', true, true, true, true),
  (1, 'Справочник бригад', '/team', true, true, true, true),
  (1, 'Справочник сотрудников', '/worker', true, true, true, true),
  (1, 'Справочник серийных номеров', '/serial-number', true, true, true, true),
  (1, 'Местоположение метриала', '/material-location', true, true, true, true),
  (1, 'Материала привязанные к накладной', '/invoice-materials', true, true, true, true),
  (1, 'Бракованные материлы', '/material-defect', true, true, true, true),
  (1, 'Корректировка оператора', '/invoie/correction', true, true, true, true)
) AS values_tobe_inserted(role_id, resource_name, resource_url, r, w, u, d)
WHERE NOT EXISTS (
  SELECT *
  FROM permissions
  WHERE  
    permissions.role_id = values_tobe_inserted.role_id
    AND permissions.resource_url = values_tobe_inserted.resource_url
    AND permissions.resource_name = values_tobe_inserted.resource_name  
    AND permissions.r = values_tobe_inserted.r
    AND permissions.w = values_tobe_inserted.w
    AND permissions.u = values_tobe_inserted.u
    AND permissions.u = values_tobe_inserted.d
);

