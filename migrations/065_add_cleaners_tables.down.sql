-- Удаление таблиц уборщиц (в обратном порядке из-за внешних ключей)
DROP TABLE IF EXISTS cleaner_apartments CASCADE;
DROP TABLE IF EXISTS cleaners CASCADE;
