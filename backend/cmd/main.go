package main

import (
	"backend/internal/app"
	_ "backend/docs"
	"flag"
)

var (
	configPath = flag.String("config", "./config/config.yaml", "path to config file. default: ./config/config.yaml")
)

// @title						BodyFuel API
// @version					1.0
// @description				REST API для фитнес-приложения BodyFuel. Предоставляет управление пользователями, тренировками, питанием и персонализированными рекомендациями.
// @contact.name				Danila Maslov
// @license.name				MIT
//
// @host						localhost:8080
// @BasePath					/api/v1
// @schemes					http https
//
// @tag.name					Auth
// @tag.description			Регистрация, вход, refresh-токены, верификация email/телефона, восстановление пароля
// @tag.name					User Info
// @tag.description			Основная информация о пользователе (имя, email, телефон)
// @tag.name					User Params
// @tag.description			Физические параметры пользователя (рост, вес, цель)
// @tag.name					User Weight
// @tag.description			История записей веса пользователя
// @tag.name					User Calories
// @tag.description			Ручной трекинг калорий вне дневника питания
// @tag.name					Exercises
// @tag.description			Справочник упражнений
// @tag.name					Workouts
// @tag.description			Тренировки: генерация, история, управление
// @tag.name					Workout Exercises
// @tag.description			Управление упражнениями внутри конкретной тренировки
// @tag.name					Nutrition
// @tag.description			Дневник питания: анализ фото через AI, записи о еде, отчёты
// @tag.name					Recommendations
// @tag.description			Персонализированные рекомендации, сгенерированные GPT на основе профиля
// @tag.name					Devices
// @tag.description			Регистрация device-токенов для APNs push-уведомлений
// @tag.name					Photo
// @tag.description			Загрузка аватаров через presigned URL (MinIO/S3)
// @tag.name					Tasks
// @tag.description			Очередь фоновых задач executor'а (email, SMS, push)
//
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Введите "Bearer <access_token>". Токен выдаётся при входе (POST /auth/login) или обновлении (POST /auth/refresh).
func main() {
	flag.Parse()

	application := app.NewApp(*configPath)
	application.Run()
}
