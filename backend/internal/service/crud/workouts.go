package crud

//workout с параметрами
//(15 - 30 - 1ч - 1.30ч - 2ч.)
//Уровень нагрузки (слабый,умеренный, сильный),
//Параметры человека(вес, калории, возраст, тип жизни, какая уже тренировка за неделю, на какой тип мышц) ->
//
//	отправляет это в executor ->
//	executor отдает recommendation ->
//	recommendation отдает open ai -> составляется тренировка из типа тренировки (кардио, силовая) ->
//	отдается массив из упражнений, а также, сколько минут на каждое, дальше есть кнопка начать, где каждые 10 минут собираются данные с пульсометра.
//
//	упражненение - таблица следующего вида:
//
//Exercise
//		ui упражнения
//		level_training
//		name - название упражнения
//	    (cardio, upper_body, lower_body, full_body, flexibility) тип упражнения
//       (описание упражнения) text
//		base_count_reps()
//		steps
//		(link gif)
//		(street, gym, home) place_exersize
//		(avg_calories_per)
//		base_relax_time
//
//
//		"beginner"
//		"push-ups"
//		"upper_body"
//		"TEXT"
//		"20"
//		"null"
//		"home"
//		"0,3"
//		"3"
//
//       "beginner"
//		"pull_ups"
//		"upper_body"
//		"TEXT"
//		"5"
//		"null"
//		"1"
//		"4"
//
//		""
//
//
//
//
