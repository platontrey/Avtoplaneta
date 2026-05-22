/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

export default function OperatorInstructions() {
  return (
    <div className="max-w-4xl mx-auto px-6 py-8">
      <h1 className="text-3xl font-bold mb-6">Инструкция для оператора</h1>

      <div className="space-y-8">
        <section>
          <h2 className="text-2xl font-semibold mb-4">Обзор роли</h2>
          <p className="text-gray-700">
            Оператор отвечает за ежедневное управление инвентарем автозапчастей. Ваши основные обязанности включают ведение каталога запчастей, ввод и обновление данных о запчастях, а также контроль качества информации.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Разрешения</h2>
          <ul className="list-disc list-inside space-y-1">
            <li>✅ Создание новых запчастей</li>
            <li>✅ Просмотр списка запчастей</li>
            <li>✅ Обновление информации о существующих запчастях</li>
            <li>❌ Удаление запчастей</li>
            <li>❌ Доступ к заказам</li>
            <li>❌ Управление пользователями</li>
            <li>❌ Администрирование системы</li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Работа в веб-интерфейсе</h2>

          <h3 className="text-xl font-medium mb-3">Вход в систему</h3>
          <ol className="list-decimal list-inside space-y-2 mb-4">
            <li>Откройте браузер и перейдите на адрес системы Avtoplaneta</li>
            <li>Введите ваш логин и пароль</li>
            <li>Нажмите "Войти"</li>
          </ol>

          <h3 className="text-xl font-medium mb-3">Просмотр инвентаря</h3>
          <ol className="list-decimal list-inside space-y-2 mb-4">
            <li>После входа вы увидите главную страницу с инвентарем</li>
            <li>Используйте поисковую строку для поиска запчастей по названию, артикулу или описанию</li>
            <li>Применяйте фильтры по категориям для сужения списка</li>
          </ol>

          <h3 className="text-xl font-medium mb-3">Добавление новой запчасти</h3>
          <ol className="list-decimal list-inside space-y-2 mb-4">
            <li>Нажмите кнопку "Добавить запчасть" или перейдите в раздел "Инвентарь"</li>
            <li>Заполните форму:
              <ul className="list-disc list-inside ml-6 mt-2 space-y-1">
                <li>Название запчасти</li>
                <li>Артикул (уникальный код)</li>
                <li>Категория</li>
                <li>Описание</li>
                <li>Цена</li>
                <li>Количество на складе</li>
              </ul>
            </li>
            <li>Загрузите фотографии запчасти (рекомендуется несколько ракурсов)</li>
            <li>Нажмите "Сохранить"</li>
          </ol>

          <h3 className="text-xl font-medium mb-3">Обновление информации о запчасти</h3>
          <ol className="list-decimal list-inside space-y-2 mb-4">
            <li>Найдите нужную запчасть в списке или через поиск</li>
            <li>Нажмите на карточку запчасти для просмотра деталей</li>
            <li>Нажмите кнопку "Редактировать"</li>
            <li>Внесите необходимые изменения</li>
            <li>Сохраните изменения</li>
          </ol>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Работа в мобильном приложении</h2>

          <h3 className="text-xl font-medium mb-3">Установка и вход</h3>
          <ol className="list-decimal list-inside space-y-2 mb-4">
            <li>Скачайте приложение Avtoplaneta из Google Play</li>
            <li>Запустите приложение</li>
            <li>Введите логин и пароль для входа</li>
          </ol>

          <h3 className="text-xl font-medium mb-3">Просмотр инвентаря</h3>
          <ol className="list-decimal list-inside space-y-2 mb-4">
            <li>После входа откроется главный экран с инвентарем</li>
            <li>Используйте поиск для быстрого нахождения запчастей</li>
            <li>Просматривайте детали запчастей, нажимая на карточки</li>
          </ol>

          <h3 className="text-xl font-medium mb-3">Добавление запчасти</h3>
          <ol className="list-decimal list-inside space-y-2 mb-4">
            <li>Нажмите кнопку "+" или "Добавить" в нижнем меню</li>
            <li>Выберите "Добавить запчасть"</li>
            <li>Заполните информацию</li>
            <li>Сделайте фото запчасти с помощью камеры устройства</li>
            <li>Сохраните новую запчасть</li>
          </ol>

          <h3 className="text-xl font-medium mb-3">Обновление запчасти</h3>
          <ol className="list-decimal list-inside space-y-2 mb-4">
            <li>Найдите запчасть в списке</li>
            <li>Нажмите и удерживайте для открытия меню действий</li>
            <li>Выберите "Редактировать"</li>
            <li>Внесите изменения</li>
            <li>Сохраните</li>
          </ol>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Лучшие практики</h2>

          <h3 className="text-xl font-medium mb-3">Качество данных</h3>
          <ul className="list-disc list-inside space-y-1 mb-4">
            <li>Всегда проверяйте правильность введенной информации</li>
            <li>Используйте четкие фотографии высокого качества</li>
            <li>Заполняйте все обязательные поля</li>
            <li>Проверяйте уникальность артикулов</li>
          </ul>

          <h3 className="text-xl font-medium mb-3">Организация работы</h3>
          <ul className="list-disc list-inside space-y-1 mb-4">
            <li>Регулярно обновляйте информацию о наличии запчастей</li>
            <li>Категоризируйте запчасти правильно для удобного поиска</li>
            <li>Ведите подробные описания для сложных запчастей</li>
          </ul>

          <h3 className="text-xl font-medium mb-3">Безопасность</h3>
          <ul className="list-disc list-inside space-y-1">
            <li>Не делитесь своими учетными данными</li>
            <li>Выходите из системы при завершении работы</li>
            <li>Сообщайте о любых подозрительных действиях</li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Частые вопросы</h2>
          <div className="space-y-4">
            <div>
              <h4 className="font-medium">Почему я не вижу кнопку "Удалить" у запчастей?</h4>
              <p className="text-gray-700">У операторов нет прав на удаление запчастей. Обратитесь к менеджеру или администратору для удаления.</p>
            </div>
            <div>
              <h4 className="font-medium">Как добавить несколько фото к запчасти?</h4>
              <p className="text-gray-700">В веб-интерфейсе используйте кнопку "Добавить фото" несколько раз. В мобильном приложении сделайте несколько снимков подряд.</p>
            </div>
            <div>
              <h4 className="font-medium">Что делать, если артикул уже существует?</h4>
              <p className="text-gray-700">Каждый артикул должен быть уникальным. Проверьте правильность ввода или обратитесь к администратору для проверки существующих данных.</p>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-semibold mb-4">Контакты для поддержки</h2>
          <p className="text-gray-700">При возникновении проблем обращайтесь к вашему менеджеру или администратору системы.</p>
        </section>
      </div>
    </div>
  );
}