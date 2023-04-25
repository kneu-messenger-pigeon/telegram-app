{{.NamePrefix}} {{.Name}}, Ваша загальна успішність у навчанні:

{{range $index, $d := .Disciplines}}
{{incr $index}}. {{$d.Discipline.Name}}
     *результат {{$d.ScoreRating.Total}}*, _рейтинг #{{$d.ScoreRating.Rating}}/{{$d.ScoreRating.StudentsCount}}_
{{else}}
Навчальних дисциплін ще не зареєстровано
{{end}}

Вимкнути бот - /reset

❗Увага❗
Перевіряйте оцінки в [офіційному журналі успішності КНЕУ](https://cutt.ly/Dekanat)
Цей Бот не є офіційним джерелом даних про успішність.

