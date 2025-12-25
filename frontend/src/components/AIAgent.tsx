/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

// Type declarations for Web Speech API
declare global {
  interface Window {
    SpeechRecognition: typeof SpeechRecognition;
    webkitSpeechRecognition: typeof SpeechRecognition;
  }
}

interface SpeechRecognition extends EventTarget {
  continuous: boolean;
  interimResults: boolean;
  lang: string;
  start(): void;
  stop(): void;
  abort(): void;
  onstart: ((this: SpeechRecognition, ev: Event) => any) | null;
  onresult: ((this: SpeechRecognition, ev: SpeechRecognitionEvent) => any) | null;
  onend: ((this: SpeechRecognition, ev: Event) => any) | null;
  onerror: ((this: SpeechRecognition, ev: SpeechRecognitionErrorEvent) => any) | null;
}

interface SpeechRecognitionEvent extends Event {
  results: SpeechRecognitionResultList;
  resultIndex: number;
}

interface SpeechRecognitionErrorEvent extends Event {
  error: string;
  message: string;
}

interface SpeechRecognitionResultList {
  readonly length: number;
  item(index: number): SpeechRecognitionResult;
  [index: number]: SpeechRecognitionResult;
}

interface SpeechRecognitionResult {
  readonly length: number;
  item(index: number): SpeechRecognitionAlternative;
  [index: number]: SpeechRecognitionAlternative;
  isFinal: boolean;
}

interface SpeechRecognitionAlternative {
  transcript: string;
  confidence: number;
}

declare var SpeechRecognition: {
  prototype: SpeechRecognition;
  new(): SpeechRecognition;
};

import { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Bot, X, Mic, MicOff, Send } from 'lucide-react';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Badge } from './ui/badge';
import { motion, AnimatePresence } from 'framer-motion';
import { aiAgentApi } from '../lib/api';

interface Message {
  id: string;
  text: string;
  sender: 'user' | 'ai';
  timestamp: Date;
  type?: 'text' | 'action' | 'error';
  action?: any; // Для хранения контекста действий
}

function AIAgent({ isOpen, onToggle }: { isOpen: boolean; onToggle: () => void }) {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: '1',
      text: 'Привет! Я ваш ИИ-помощник. Чем могу помочь?',
      sender: 'ai',
      timestamp: new Date(),
      type: 'text'
    }
  ]);
  const [inputValue, setInputValue] = useState('');
  const [isListening, setIsListening] = useState(false);
  const [isTyping, setIsTyping] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const recognitionRef = useRef<SpeechRecognition | null>(null);

  const navigate = useNavigate();

  // Функция выполнения действий
  const executeAction = async (action: any) => {
    if (!action || !action.type) return;

    try {
      switch (action.type) {
        case 'navigate':
          if (action.path) {
            navigate(action.path);
            addMessage(`Переход на страницу: ${action.path}`, 'ai', 'action');
          }
          break;

        case 'search':
          if (action.query) {
            // Выполняем поиск через API
            const results = await aiAgentApi.searchParts(action.query);
            if (results && results.length > 0) {
              addMessage(`Найдено ${results.length} запчастей по запросу "${action.query}"`, 'ai', 'action');
              // Можно показать результаты или перейти к странице поиска
              navigate('/inventory');
            } else {
              addMessage(`По запросу "${action.query}" ничего не найдено`, 'ai', 'action');
            }
          }
          break;

        case 'search_and_delete':
          if (action.part_name) {
            // Сначала ищем запчасть
            const results = await aiAgentApi.searchParts(action.part_name);
            if (results && results.length === 1) {
              // Если найдена ровно одна запчасть, удаляем её
              await aiAgentApi.deletePart(results[0].id);
              addMessage(`Запчасть "${results[0].name}" успешно удалена`, 'ai', 'action');
            } else if (results && results.length > 1) {
              // Если найдено несколько, предлагаем уточнить
              const options = results.slice(0, 3).map((part: any) => part.name).join(', ');
              addMessage(`Найдено несколько запчастей: ${options}. Уточните, какую именно удалить.`, 'ai', 'action');
            } else {
              addMessage(`Запчасть "${action.part_name}" не найдена`, 'ai', 'action');
            }
          }
          break;

        case 'search_and_update':
          if (action.part_name && action.field && action.value) {
            // Сначала ищем запчасть
            const results = await aiAgentApi.searchParts(action.part_name);
            if (results && results.length === 1) {
              // Если найдена ровно одна запчасть, обновляем её
              const updateData: any = {};
              updateData[action.field] = action.value;
              await aiAgentApi.updatePart(results[0].id, updateData);
              addMessage(`Запчасть "${results[0].name}" успешно обновлена`, 'ai', 'action');
            } else if (results && results.length > 1) {
              // Если найдено несколько, предлагаем уточнить
              const options = results.slice(0, 3).map((part: any) => part.name).join(', ');
              addMessage(`Найдено несколько запчастей: ${options}. Уточните, какую именно обновить.`, 'ai', 'action');
            } else {
              addMessage(`Запчасть "${action.part_name}" не найдена`, 'ai', 'action');
            }
          }
          break;

        case 'add_part':
          if (action.data) {
            console.log('AIAgent: Выполняю add_part с данными:', action.data);
            try {
              await aiAgentApi.addPart(action.data);
              console.log('AIAgent: add_part выполнен успешно');
              addMessage(`Запчасть "${action.data.name}" успешно добавлена! Цена: ${action.data.price || 0} руб., количество: ${action.data.quantity || 1} шт.`, 'ai', 'action');
            } catch (error) {
              console.error('AIAgent: Ошибка при добавлении запчасти:', error);
              addMessage('Произошла ошибка при добавлении запчасти. Попробуйте еще раз.', 'ai', 'error');
            }
          } else {
            console.log('AIAgent: action.data отсутствует для add_part');
          }
          break;

        case 'update_part':
          if (action.id && action.data) {
            await aiAgentApi.updatePart(action.id, action.data);
            addMessage(`Запчасть с ID ${action.id} успешно обновлена`, 'ai', 'action');
          }
          break;

        case 'delete_part':
          if (action.id) {
            await aiAgentApi.deletePart(action.id);
            addMessage(`Запчасть с ID ${action.id} успешно удалена`, 'ai', 'action');
          }
          break;

        case 'clarify':
          if (action.options && Array.isArray(action.options)) {
            const optionsText = action.options.join(', ');
            addMessage(`Уточните выбор: ${optionsText}`, 'ai', 'action');
            // Сохраняем контекст для следующего ответа пользователя
            setMessages(prev => prev.map(msg =>
              msg.id === Date.now().toString() ? {...msg, action: action} : msg
            ));
          }
          break;

        default:
          console.log('Неизвестное действие:', action);
      }
    } catch (error) {
      console.error('Ошибка выполнения действия:', error);
      addMessage('Произошла ошибка при выполнении действия. Попробуйте еще раз.', 'ai', 'error');
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // Web Speech API для голосового ввода
  const startVoiceInput = () => {
    if (!('webkitSpeechRecognition' in window) && !('SpeechRecognition' in window)) {
      addMessage('Голосовой ввод не поддерживается в вашем браузере', 'ai', 'error');
      return;
    }

    const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
    recognitionRef.current = new SpeechRecognition();

    recognitionRef.current.lang = 'ru-RU';
    recognitionRef.current.continuous = false;
    recognitionRef.current.interimResults = false;

    recognitionRef.current.onstart = () => {
      setIsListening(true);
    };

    recognitionRef.current.onresult = (event: SpeechRecognitionEvent) => {
      const transcript = event.results[0][0].transcript;
      setInputValue(transcript);
      setIsListening(false);
    };

    recognitionRef.current.onend = () => {
      setIsListening(false);
    };

    recognitionRef.current.onerror = (event: SpeechRecognitionErrorEvent) => {
      setIsListening(false);
      addMessage(`Ошибка распознавания речи: ${event.error}`, 'ai', 'error');
    };

    try {
      recognitionRef.current.start();
    } catch (error) {
      setIsListening(false);
      addMessage('Не удалось начать распознавание речи', 'ai', 'error');
    }
  };

  const stopVoiceInput = () => {
    if (recognitionRef.current) {
      recognitionRef.current.stop();
    }
    setIsListening(false);
  };

  const addMessage = (text: string, sender: 'user' | 'ai', type: 'text' | 'action' | 'error' = 'text') => {
    const newMessage: Message = {
      id: Date.now().toString(),
      text,
      sender,
      timestamp: new Date(),
      type
    };
    setMessages(prev => [...prev, newMessage]);
  };

  const sendMessage = async () => {
    if (!inputValue.trim()) return;

    const userMessage = inputValue.trim();
    setInputValue('');
    addMessage(userMessage, 'user');

    setIsTyping(true);

    try {
      // Проверяем, есть ли ожидаемое уточнение
      const lastAIMessage = messages.slice().reverse().find(msg => msg.sender === 'ai' && msg.action);
      if (lastAIMessage && lastAIMessage.action && lastAIMessage.action.type === 'clarify') {
        // Обработка выбора варианта уточнения
        await handleClarificationResponse(userMessage, lastAIMessage.action);
      } else {
        // Отправка обычного сообщения на backend для обработки ИИ
        const result = await aiAgentApi.chat(userMessage, {
          currentPage: window.location.pathname,
          userAgent: navigator.userAgent,
          timestamp: new Date().toISOString(),
          messages: messages.slice(-5), // последние 5 сообщений для контекста
          sessionId: `session_${Date.now()}`, // простой ID сессии
          language: navigator.language
        });

        // Добавляем ответ ИИ
        addMessage(result.response, 'ai');

        // Если есть действие для выполнения
        console.log('AIAgent: Получен ответ от сервера:', result);
        if (result.action) {
          console.log('AIAgent: Найдено действие для выполнения:', result.action);
          executeAction(result.action);
        } else {
          console.log('AIAgent: Действие не найдено в ответе');
        }
      }

    } catch (error) {
      console.error('Ошибка при отправке сообщения:', error);
      addMessage('Извините, произошла ошибка при обработке вашего запроса. Попробуйте еще раз.', 'ai', 'error');
    } finally {
      setIsTyping(false);
    }
  };

  // Обработка ответа на уточнение
  const handleClarificationResponse = async (userResponse: string, clarificationAction: any) => {
    try {
      const options = clarificationAction.options || [];
      const selectedIndex = parseInt(userResponse) - 1; // Предполагаем, что пользователь ввел номер

      let selectedOption: string | null = null;

      if (selectedIndex >= 0 && selectedIndex < options.length) {
        selectedOption = options[selectedIndex];
      } else {
        // Ищем точное совпадение
        for (const option of options) {
          if (userResponse.toLowerCase().includes(option.toLowerCase()) ||
              option.toLowerCase().includes(userResponse.toLowerCase())) {
            selectedOption = option;
            break;
          }
        }
      }

      if (selectedOption) {
        addMessage(`Вы выбрали: ${selectedOption}`, 'ai', 'action');

        // Выполняем действие на основе типа уточнения
        if (clarificationAction.action === 'delete') {
          // Нужно найти ID запчасти по имени
          const parts = await aiAgentApi.searchParts(selectedOption);
          if (parts && parts.length > 0) {
            await aiAgentApi.deletePart(parts[0].id);
            addMessage(`Запчасть "${selectedOption}" успешно удалена`, 'ai', 'action');
          }
        }
        // Можно добавить другие типы действий
      } else {
        addMessage(`Не удалось распознать выбор. Попробуйте указать номер или название из списка.`, 'ai', 'error');
      }
    } catch (error) {
      console.error('Ошибка при обработке уточнения:', error);
      addMessage('Произошла ошибка при обработке вашего выбора.', 'ai', 'error');
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  return (
    <>
      {/* Плавающая кнопка */}
      <motion.div
        className="fixed bottom-4 right-4 z-50 hidden md:block"
        initial={{ scale: 0 }}
        animate={{ scale: 1 }}
        transition={{ type: 'spring', stiffness: 260, damping: 20 }}
      >
        <Button
          onClick={onToggle}
          className="rounded-full w-14 h-14 shadow-lg hover:shadow-xl transition-shadow"
          size="lg"
        >
          {isOpen ? (
            <X className="w-6 h-6" />
          ) : (
            <Bot className="w-6 h-6" />
          )}
        </Button>
      </motion.div>

      {/* Чат виджет */}
      <AnimatePresence>
        {isOpen && (
          <motion.div
            className="fixed md:bottom-20 md:right-4 bottom-0 right-0 z-40"
            initial={{ opacity: 0, scale: 0.8, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.8, y: 20 }}
            transition={{ type: 'spring', stiffness: 300, damping: 30 }}
          >
            <Card className="w-screen h-screen md:w-80 md:max-w-[calc(100vw-2rem)] md:h-96 md:max-h-[calc(100vh-8rem)] shadow-2xl border-2">
              <CardHeader className="pb-3 flex-shrink-0">
                <CardTitle className="flex items-center gap-2 text-lg">
                  <Bot className="w-5 h-5 text-blue-500 flex-shrink-0" />
                  <span className="truncate">ИИ-помощник</span>
                  <Badge variant="secondary" className="text-xs flex-shrink-0">
                    Avtoplaneta
                  </Badge>
                </CardTitle>
              </CardHeader>

              <CardContent className="flex flex-col h-[calc(100%-4rem)] p-0 min-h-0">
                {/* Сообщения */}
                <div className="flex-1 overflow-y-auto p-4 space-y-3 min-h-0">
                  {messages.map((message) => (
                    <motion.div
                      key={message.id}
                      className={`flex ${message.sender === 'user' ? 'justify-end' : 'justify-start'}`}
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ duration: 0.3 }}
                    >
                      <div
                        className={`max-w-[80%] p-3 rounded-lg text-sm ${
                          message.sender === 'user'
                            ? 'bg-blue-500 text-white'
                            : message.type === 'error'
                            ? 'bg-red-100 text-red-800 border border-red-200'
                            : 'bg-gray-100 text-gray-800'
                        }`}
                      >
                        {message.text}
                        <div className="text-xs opacity-70 mt-1">
                          {message.timestamp.toLocaleTimeString('ru-RU', {
                            hour: '2-digit',
                            minute: '2-digit'
                          })}
                        </div>
                      </div>
                    </motion.div>
                  ))}

                  {/* Индикатор печати */}
                  {isTyping && (
                    <motion.div
                      className="flex justify-start"
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                    >
                      <div className="bg-gray-100 p-3 rounded-lg">
                        <div className="flex space-x-1">
                          <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                          <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
                          <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
                        </div>
                      </div>
                    </motion.div>
                  )}

                  <div ref={messagesEndRef} />
                </div>

                {/* Ввод сообщения */}
                <div className="border-t p-3 flex-shrink-0">
                  <div className="flex gap-2 items-end">
                    <Input
                      value={inputValue}
                      onChange={(e) => setInputValue(e.target.value)}
                      onKeyPress={handleKeyPress}
                      placeholder="Введите сообщение..."
                      className="flex-1 min-w-0"
                      disabled={isTyping}
                    />

                    {/* Кнопка голосового ввода */}
                    <Button
                      onClick={isListening ? stopVoiceInput : startVoiceInput}
                      variant="outline"
                      size="sm"
                      className={`shrink-0 w-8 h-8 p-0 ${isListening ? 'bg-red-100 border-red-300' : ''}`}
                      disabled={isTyping}
                    >
                      {isListening ? (
                        <MicOff className="w-4 h-4 text-red-500" />
                      ) : (
                        <Mic className="w-4 h-4" />
                      )}
                    </Button>

                    {/* Кнопка отправки */}
                    <Button
                      onClick={sendMessage}
                      size="sm"
                      className="shrink-0 w-8 h-8 p-0"
                      disabled={!inputValue.trim() || isTyping}
                    >
                      <Send className="w-4 h-4" />
                    </Button>
                  </div>

                  {isListening && (
                    <div className="text-xs text-red-600 mt-2 flex items-center gap-1">
                      <span>🎤</span>
                      Слушаю...
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  );
}

export default AIAgent;