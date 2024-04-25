INSERT INTO resources(category, name, url)
SELECT *
FROM (VALUES
  ('Накладные', 'Накладная приход', '/input'),
  ('Накладные', 'Накладная отпуск', '/output'),
  ('Накладные', 'Накладная возврат', '/return'),
  ('Накладные', 'Накладная списание', '/write-off'),
  ('Накладные', 'Накладная объект', '/invoice-object'),
  ('Накладные', 'Корректировка оператора', '/correction'),
  ('Накладные', 'Материала привязанные к накладной', '/invoice-materials'),
  ('Администратирование', 'администрирование пользователями', '/user'),
  ('Администратирование', 'администрирование действия пользователей', '/user-action'),
  ('Администратирование', 'администрирование доступами пользователей в проекты', '/user-in-projects'),
  ('Администратирование', 'Администрирование ресурсами', '/resource'),
  ('Администратирование', 'Администрирование ролями', '/role'),
  ('Администратирование', 'Администрирование доступами', '/permission'),
  ('Справочник', 'Справочник материалов', '/material'),
  ('Справочник', 'Ценники материалов', '/material-cost'),
  ('Справочник', 'Справочник проектов', '/project'),
  ('Справочник', 'Справочник районов', '/district'),
  ('Справочник', 'Справочник сервисов', '/operation'),
  ('Справочник', 'Справочник объектов', '/object'),
  ('Справочник', 'Справочник бригад', '/team'),
  ('Справочник', 'Справочник сотрудников', '/worker'),
  ('Справочник', 'Справочник серийных номеров', '/serial-number'),
  ('Справочник', 'Местоположение метриала', '/material-location'),
  ('Справочник', 'Бракованные материлы', '/material-defect')
) AS values_tobe_inserted(category, name, url)
WHERE NOT EXISTS (
  SELECT *
  FROM resources
  WHERE  
    resources.name = values_tobe_inserted.name
    AND resources.url = values_tobe_inserted.url
    AND resources.category = values_tobe_inserted.category
);

