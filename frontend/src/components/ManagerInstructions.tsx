/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

export default function ManagerInstructions() {
  return (
    <div className="max-w-4xl mx-auto px-6 py-8">
      <h1 className="text-3xl font-bold mb-6">Инструкция для менеджера</h1>

      <div className="space-y-8">
        <section>
          <h2 className="text-2xl font-semibold mb-4">Обзор роли</h2>
          <p className="text-gray-700">
            Менеджер отвечает за управление заказами, надзор за инвентарем и координацию работы команды. Ваши обязанности включают обработку заказов, управление отношениями с клиентами и контроль за состоянием склада.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Разрешения</h2>
          <ul className="list-disc list-inside space-y-1">
            <li>✅ Создание и удаление заказов</li>
            <li>✅ Просмотр информации о пользователях (только чтение)</li>
            <li>✅ Полный CRUD доступ к запчастям (создание, чтение, обновление, удаление)</li>
            <li>❌ Нет доступа к статусу сервера и логам</li>
            <li>❌ Не может создавать или удалять учетные записи пользователей</li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Управление заказами</h2>

          <h3 className="text-xl font-medium mb-3">Просмотр заказов</h3>
          <ol className="list-decimal list-inside space-y-2 mb-4">
            <li>Перейдите в раздел "Заказы" в главном меню</li>
            <li>Просмотрите список всех заказов с цветовой кодировкой статусов:
              <ul className="list-disc list-inside ml-6 mt-2 space-y-1">
                <li>Красный: Новый заказ</li>
                <li>Коричневый: В обработке</li>
                <li>Желтый: Готов к выдаче</li>
                <li>Зеленый: Выполнен</li>
              </ul>
            </li>
          </ol>

          <h3 className="text-xl font-medium mb-3">Создание нового заказа</h3>
          <ol className="list-decimal list-inside space-y-2 mb-4">
            <li>Нажмите кнопку "Создать заказ"</li>
            <li>Выберите клиента или введите информацию о новом клиенте</li>
            <li>Добавьте запчасти в заказ:
              <ul className="list-disc list-inside ml-6 mt-2 space-y-1">
                <li>Используйте поиск для нахождения нужных запчастей</li>
                <li>Укажите количество для каждой позиции</li>
                <li>Проверьте наличие на складе</li>
              </ul>
            </li>
            <li>Установите приоритет и сроки выполнения</li>
            <li>Сохраните заказ</li>
          </ol>

          <h3 className="text-xl font-medium mb-3">Обновление статуса заказа</h3>
          <ol className="list-decimal list-inside space-y-2 mb-4">
            <li>Откройте детали заказа</li>
            <li>Измените статус в соответствии с прогрессом</li>
            <li>Добавьте комментарии о выполненных действиях</li>
          </ol>

          <h3 className="text-xl font-medium mb-3">Удаление заказа</h3>
          <ol className="list-decimal list-inside space-y-2 mb-4">
            <li>Откройте заказ</li>
            <li>Нажмите кнопку "Удалить заказ"</li>
            <li>Подтвердите удаление (только если заказ не выполнен)</li>
          </ol>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Управление инвентарем</h2>

          <h3 className="text-xl font-medium mb-3">Просмотр и управление запчастями</h3>
          <ol className="list-decimal list-inside space-y-2 mb-4">
            <li>Перейдите в раздел "Инвентарь"</li>
            <li>Используйте фильтры и поиск для навигации</li>
            <li>Для каждой запчасти доступны действия:
              <ul className="list-disc list-inside ml-6 mt-2 space-y-1">
                <li>Просмотр деталей</li>
                <li>Редактирование информации</li>
                <li>Удаление (с подтверждением)</li>
              </ul>
            </li>
          </ol>

          <h3 className="text-xl font-medium mb-3">Добавление запчастей</h3>
          <ol className="list-decimal list-inside space-y-2 mb-4">
            <li>Нажмите "Добавить запчасть"</li>
            <li>Заполните полную информацию</li>
            <li>Загрузите качественные фотографии</li>
            <li>Установите начальное количество</li>
          </ol>

          <h3 className="text-xl font-medium mb-3">Удаление запчастей</h3>
          <ol className="list-decimal list-inside space-y-2 mb-4">
            <li>Найдите запчасть</li>
            <li>Нажмите "Удалить"</li>
            <li>Подтвердите действие</li>
          </ol>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Аналитика и отчеты</h2>
          <ol className="list-decimal list-inside space-y-2">
            <li>Перейдите в раздел "Аналитика"</li>
            <li>Просмотрите статистику по категориям запчастей</li>
            <li>Мониторьте объем продаж</li>
            <li>Следите за состоянием склада</li>
            <li>Анализируйте активные заказы</li>
          </ol>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Лучшие практики</h2>

          <h3 className="text-xl font-medium mb-3">Управление заказами</h3>
          <ul className="list-disc list-inside space-y-1 mb-4">
            <li>Регулярно обновляйте статусы заказов</li>
            <li>Проверяйте наличие запчастей перед подтверждением заказа</li>
            <li>Коммуницируйте с клиентами о сроках выполнения</li>
            <li>Ведите подробные комментарии к заказам</li>
          </ul>

          <h3 className="text-xl font-medium mb-3">Контроль инвентаря</h3>
          <ul className="list-disc list-inside space-y-1 mb-4">
            <li>Поддерживайте актуальную информацию о наличии</li>
            <li>Регулярно проводите инвентаризацию</li>
            <li>Удаляйте только ненужные или устаревшие запчасти</li>
            <li>Следите за качеством фотографий и описаний</li>
          </ul>

          <h3 className="text-xl font-medium mb-3">Координация команды</h3>
          <ul className="list-disc list-inside space-y-1">
            <li>Контролируйте работу операторов</li>
            <li>Распределяйте задачи по управлению инвентарем</li>
            <li>Обеспечивайте своевременное выполнение заказов</li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Частые вопросы</h2>
          <div className="space-y-4">
            <div>
              <h4 className="font-medium">Как изменить статус заказа?</h4>
              <p className="text-gray-700">Откройте заказ и выберите новый статус из выпадающего списка. Сохраните изменения.</p>
            </div>
            <div>
              <h4 className="font-medium">Могу ли я удалить запчасть, которая есть в активном заказе?</h4>
              <p className="text-gray-700">Нет, система не позволит удалить запчасть, используемую в незавершенных заказах. Сначала завершите или измените заказ.</p>
            </div>
            <div>
              <h4 className="font-medium">Как просмотреть информацию о пользователях?</h4>
              <p className="text-gray-700">Перейдите в раздел "Пользователи" - у вас есть доступ только на чтение для просмотра списка пользователей и их ролей.</p>
            </div>
            <div>
              <h4 className="font-medium">Что делать при нехватке запчастей на складе?</h4>
              <p className="text-gray-700">Отметьте заказ как "Ожидает поставки" и свяжитесь с поставщиками. Обновите информацию о наличии после поступления.</p>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Контакты для поддержки</h2>
          <p className="text-gray-700">При технических проблемах обращайтесь к администратору системы. Для вопросов по бизнес-процессам консультируйтесь с руководством компании.</p>
        </section>
      </div>
    </div>
  );
}