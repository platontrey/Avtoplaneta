import { useQuery } from '@tanstack/react-query'

import { API_BASE_URL } from '@/lib/api'
import type { User } from '../types'

/**
 * Единый запрос текущего пользователя.
 *
 * Раньше каждый хук авторизации держал собственный useState и дёргал /auth/me
 * из useEffect при монтировании. PartBlock вызывает useAuth, а PartBlock
 * рисуется на каждую запчасть — то есть список из пятидесяти позиций отправлял
 * пятьдесят одинаковых запросов, и кнопки «изменить» и «удалить» появлялись
 * по мере того, как каждая строка дожидалась своего ответа.
 *
 * Общий ключ react-query решает это: сколько бы компонентов ни спросило
 * пользователя, запрос уйдёт один, а остальные получат готовые данные.
 */
export const AUTH_ME_QUERY_KEY = ['auth', 'me'] as const

const withInitials = (user: User): User => ({
  ...user,
  initials: user.name
    ? user.name
        .split(' ')
        .map((part) => part[0])
        .join('')
        .toUpperCase()
    : '',
})

export const fetchCurrentUser = async (): Promise<User | null> => {
  try {
    const response = await fetch(`${API_BASE_URL}/auth/me`, {
      credentials: 'include',
    })
    if (!response.ok) {
      // Неавторизован — это не ошибка загрузки, а обычное состояние.
      return null
    }
    return withInitials((await response.json()) as User)
  } catch {
    return null
  }
}

export const useCurrentUser = () =>
  useQuery({
    queryKey: AUTH_ME_QUERY_KEY,
    queryFn: fetchCurrentUser,
    // Роль меняется редко, а спрашивают её почти все экраны.
    staleTime: 5 * 60 * 1000,
    gcTime: 30 * 60 * 1000,
    retry: false,
    refetchOnWindowFocus: false,
  })
