/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import type { User } from '../features/auth/types';

interface ReadmeProps {
  user: User;
}

export default function Readme({ user }: ReadmeProps) {
  console.log('User:', user); // Temporary log to use the user prop
  return (
    <div className="max-w-4xl mx-auto px-6 py-8">
      <h1 className="text-3xl font-bold mb-6">Роли пользователей и разрешения</h1>

      <div className="space-y-8">
        <section>
          <h2 className="text-2xl font-semibold mb-4">Обзор</h2>
          <p className="text-muted-foreground">
            Система реализует ролевую систему контроля доступа (RBAC) с тремя различными ролями пользователей:
            Администратор, Менеджер и Оператор. Каждая роль имеет специфические разрешения и возможности
            в системе управления инвентарем.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Роли пользователей</h2>

          <div className="space-y-6">
            <div className="border border-border bg-card rounded-lg p-6">
              <h3 className="text-xl font-medium mb-2">1. Администратор (Admin)</h3>
              <p className="text-sm text-muted-foreground mb-3">Полный доступ к системе и контроль</p>

              <h4 className="font-medium mb-2">Разрешения:</h4>
              <ul className="list-disc list-inside space-y-1 text-sm mb-4">
                <li>✅ Создание, чтение, обновление и удаление пользователей</li>
                <li>✅ Создание, чтение, обновление и удаление запчастей</li>
                <li>✅ Создание и удаление заказов</li>
                <li>✅ Просмотр статуса сервера и логов</li>
                <li>✅ Полный доступ к конфигурации системы</li>
              </ul>

              <h4 className="font-medium mb-2">Обязанности:</h4>
              <ul className="list-disc list-inside space-y-1 text-sm">
                <li>Администрирование и обслуживание системы</li>
                <li>Управление учетными записями пользователей</li>
                <li>Контроль безопасности</li>
                <li>Мониторинг и устранение неполадок системы</li>
              </ul>
            </div>

            <div className="border border-border bg-card rounded-lg p-6">
              <h3 className="text-xl font-medium mb-2">2. Менеджер (Manager)</h3>
              <p className="text-sm text-muted-foreground mb-3">Управление заказами и надзор за пользователями</p>

              <h4 className="font-medium mb-2">Разрешения:</h4>
              <ul className="list-disc list-inside space-y-1 text-sm mb-4">
                <li>✅ Создание и удаление заказов</li>
                <li>✅ Чтение информации о пользователях (без возможности изменения)</li>
                <li>✅ Создание, чтение, обновление и удаление запчастей</li>
                <li>❌ Нет доступа к статусу сервера/логам</li>
                <li>❌ Не может создавать или удалять учетные записи пользователей</li>
              </ul>

              <h4 className="font-medium mb-2">Обязанности:</h4>
              <ul className="list-disc list-inside space-y-1 text-sm">
                <li>Обработка и управление заказами</li>
                <li>Надзор за инвентарем</li>
                <li>Координация команды</li>
                <li>Управление отношениями с клиентами</li>
              </ul>
            </div>

            <div className="border border-border bg-card rounded-lg p-6">
              <h3 className="text-xl font-medium mb-2">3. Оператор (Operator)</h3>
              <p className="text-sm text-muted-foreground mb-3">Базовые операции с инвентарем</p>

              <h4 className="font-medium mb-2">Разрешения:</h4>
              <ul className="list-disc list-inside space-y-1 text-sm mb-4">
                <li>✅ Создание и чтение запчастей</li>
                <li>✅ Обновление существующих запчастей</li>
                <li>❌ Не может удалять запчасти</li>
                <li>❌ Нет доступа к заказам</li>
                <li>❌ Нет доступа к управлению пользователями</li>
                <li>❌ Нет доступа к администрированию сервера</li>
              </ul>

              <h4 className="font-medium mb-2">Обязанности:</h4>
              <ul className="list-disc list-inside space-y-1 text-sm">
                <li>Ежедневное управление инвентарем</li>
                <li>Ведение каталога запчастей</li>
                <li>Базовый ввод и обновление данных</li>
                <li>Контроль качества информации о запчастях</li>
              </ul>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Матрица доступа к API Endpoints</h2>
          <div className="overflow-x-auto">
            <table className="min-w-full border-collapse border border-border">
              <thead>
                <tr className="bg-muted/50">
                  <th className="border border-border px-4 py-2 text-left">Endpoint</th>
                  <th className="border border-border px-4 py-2 text-left">Метод</th>
                  <th className="border border-border px-4 py-2 text-center">Админ</th>
                  <th className="border border-border px-4 py-2 text-center">Менеджер</th>
                  <th className="border border-border px-4 py-2 text-center">Оператор</th>
                </tr>
              </thead>
              <tbody>
                <tr><td className="border border-border px-4 py-2">`/admin/users`</td><td className="border border-border px-4 py-2">GET</td><td className="border border-border px-4 py-2 text-center">✅</td><td className="border border-border px-4 py-2 text-center">✅</td><td className="border border-border px-4 py-2 text-center">❌</td></tr>
                <tr><td className="border border-border px-4 py-2">`/admin/users`</td><td className="border border-border px-4 py-2">POST</td><td className="border border-border px-4 py-2 text-center">✅</td><td className="border border-border px-4 py-2 text-center">❌</td><td className="border border-border px-4 py-2 text-center">❌</td></tr>
                <tr><td className="border border-border px-4 py-2">`/admin/users/:id`</td><td className="border border-border px-4 py-2">DELETE</td><td className="border border-border px-4 py-2 text-center">✅</td><td className="border border-border px-4 py-2 text-center">❌</td><td className="border border-border px-4 py-2 text-center">❌</td></tr>
                <tr><td className="border border-border px-4 py-2">`/admin/status`</td><td className="border border-border px-4 py-2">GET</td><td className="border border-border px-4 py-2 text-center">✅</td><td className="border border-border px-4 py-2 text-center">❌</td><td className="border border-border px-4 py-2 text-center">❌</td></tr>
                <tr><td className="border border-border px-4 py-2">`/admin/logs`</td><td className="border border-border px-4 py-2">GET</td><td className="border border-border px-4 py-2 text-center">✅</td><td className="border border-border px-4 py-2 text-center">❌</td><td className="border border-border px-4 py-2 text-center">❌</td></tr>
                <tr><td className="border border-border px-4 py-2">`/inventory`</td><td className="border border-border px-4 py-2">GET</td><td className="border border-border px-4 py-2 text-center">✅</td><td className="border border-border px-4 py-2 text-center">✅</td><td className="border border-border px-4 py-2 text-center">✅</td></tr>
                <tr><td className="border border-border px-4 py-2">`/addpart`</td><td className="border border-border px-4 py-2">POST</td><td className="border border-border px-4 py-2 text-center">✅</td><td className="border border-border px-4 py-2 text-center">✅</td><td className="border border-border px-4 py-2 text-center">✅</td></tr>
                <tr><td className="border border-border px-4 py-2">`/updatepart/:id`</td><td className="border border-border px-4 py-2">PUT</td><td className="border border-border px-4 py-2 text-center">✅</td><td className="border border-border px-4 py-2 text-center">✅</td><td className="border border-border px-4 py-2 text-center">✅</td></tr>
                <tr><td className="border border-border px-4 py-2">`/deletepart/:id`</td><td className="border border-border px-4 py-2">DELETE</td><td className="border border-border px-4 py-2 text-center">✅</td><td className="border border-border px-4 py-2 text-center">✅</td><td className="border border-border px-4 py-2 text-center">❌</td></tr>
                <tr><td className="border border-border px-4 py-2">`/orders`</td><td className="border border-border px-4 py-2">GET/POST</td><td className="border border-border px-4 py-2 text-center">✅</td><td className="border border-border px-4 py-2 text-center">✅</td><td className="border border-border px-4 py-2 text-center">❌</td></tr>
              </tbody>
            </table>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Функции безопасности</h2>
          <ul className="list-disc list-inside space-y-2">
            <li><strong>Ролевое middleware:</strong> Все endpoints защищены соответствующими проверками ролей</li>
            <li><strong>Иерархические разрешения:</strong> Высшие роли наследуют разрешения низших ролей</li>
            <li><strong>Аудит логирования:</strong> Все административные действия логируются с информацией о пользователе</li>
            <li><strong>CSRF защита:</strong> Все операции изменения состояния требуют валидных CSRF токенов</li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Начало работы</h2>
          <ol className="list-decimal list-inside space-y-2">
            <li><strong>Настройка админа:</strong> Первый пользователь должен быть создан с ролью администратора</li>
            <li><strong>Назначение ролей:</strong> Администраторы могут создавать пользователей с соответствующими ролями</li>
            <li><strong>Обучение разрешениям:</strong> Убедитесь, что пользователи понимают ограничения своих ролей</li>
            <li><strong>Регулярные аудиты:</strong> Периодически проверяйте роли пользователей и паттерны доступа</li>
          </ol>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Лучшие практики</h2>
          <ul className="list-disc list-inside space-y-2">
            <li><strong>Принцип наименьших привилегий:</strong> Назначайте минимально необходимую роль каждому пользователю</li>
            <li><strong>Разделение ролей:</strong> Разные роли для разных рабочих функций</li>
            <li><strong>Регулярные проверки:</strong> Периодически аудитируйте роли пользователей и разрешения</li>
            <li><strong>Обучение:</strong> Убедитесь, что пользователи понимают свои разрешения и обязанности</li>
          </ul>
        </section>
      </div>
    </div>
  );
}
