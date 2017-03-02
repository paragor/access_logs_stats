## Access Logs Stats

[![Go Report Card](https://goreportcard.com/badge/blackbass1988/access_logs_stats)](https://goreportcard.com/report/github.com/blackbass1988/access_logs_stats)

цель приложения - сделать штуку, которая могла бы обработать "очень много" access логов и построить по ним статистику

установка GO
------------

1) поставить в систему yum/apt/brew install go...
2) поставить через [gvm](https://github.com/moovweb/gvm#installing) (аналог rvm)
3) поставить [docker](https://www.docker.com/products/overview)

Все способы имеют права на жизнь. Например, в линукс машине у меня стоит gvm, 
а на маке - через брю.

Но я также успешно экспериментировал с докером. 

для чисто билдов я бы, возможно, посоветовал использовать докер.
Хотя, возможно, в будущем будет makefile, который будет смореть и, если нету go tools, но есть докер,
 то просто скачает образы и соберет все сам.
  
  Минус второго варианта - чтобы собрать 1.5+, необходим 1.4. По ссылке описано, как собирать 1.5+

клонирование в GOPATH
---------------------

если нет $GOPATH, то создаем
```
export GOPATH=~/go
mkdir $GOPATH/{src,bin,pkg}
```

```
git clone https://github.com/blackbass1988/access_logs_stats $GOPATH/src/github.com/blackbass1988/access_logs_stats
```

сборка
------

для кросс компиляции нужен golang 1.5+
лично я собирал на go 1.7.1


common case
```
make
```

для линукса 64 бита
```
go test ./... &&  go fmt ./... && env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w"
```


для виндуса 64 бита
```
go test ./... &&  go fmt ./... && env GOOS=windows GOARCH=amd64 go build -ldflags="-s -w"
```


для макуса 64 бита
```
go test ./... &&  go fmt ./... && env GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w"
```

есть вариант не устанавливать сам Go, а воспользоваться Docker'ом

```
docker run --rm -v "$GOPATH":$GOPATH -w $(pwd) -e GOPATH=$GOPATH -e GOOS=linux -e GOARCH=amd64 golang:1.7 go test ./... &&  go fmt ./... && go build -v -ldflags="-s -w"
```

косяк в строчке выше в том, что нужен $GOPATH, чтобы она выполнилась. 

Кстати, заметил, что бинарники для линукса получаются короче,
 если собирать их в линкусе, а не делать кросс-компиляцию (100кб после запаковки с помощью upx)

более подробный список целей сборки тут https://golang.org/doc/install/source#environment

упаковка
--------

после того, как добро собралось, можно пакнуть с использованием https://upx.github.io
не стал добалять в однострочник, потому что сначала надо поставить upx.
Пусть это остается на совести

В результате использования флагов и упаковщика получается уменьшить размер бинарника с 3.6MB до 845KB

```
upx access_logs_stats
```

использование
-------------

```
./access_logs_stats -c config.json
```

пример config.json ниже

настройка 
---------

достаточно взять за основу config.json.example или пример ниже

Описание параметров под примером

```
{
  "input": "file::foo.txt",
  "regexp": ".+HTTP\/\\d.?\\d?\\s(?P<code>\\d+)[^\"]+\"[^\"]*\" \"[^\"]*\" (?P<time>\\d{1,}\\.\\d{3})",
  "period": "10s",
  "counts": ["code", "time"],
  "aggregates": ["time"],
  "filters": [
    {
      "filter": ".+",
      "prefix": "prefix2_",
      "items": [
        {
          "field": "code",
          "metrics": ["cps_200", "cps_400", "cps_500"]
        },
        {
          "field": "time",
          "metrics": ["avg", "cent_90", "min", "max"]
        }
      ]
    }
  ],
  "output": [
    {
      "type": "console",
      "settings": {}
    },
    {
      "type": "zabbix",
      "settings": {
        "zabbix_host": "127.0.0.1",
        "zabbix_port": "1234",
        "host": "localhost.localhost"
      }
    }
  ]
}

```

|поле|описание|
|----|------|
|*input*| это точка, откуда будут читаться. Здесь может быть как файл,так и пайп, например. *Experimental: syslog:udp::515/nginx*|
|*regexp*|глобальное регулярное выражение, которое нужно, чтобы выделить _поля_ для последующих вычислений|
|*period*|период, раз во сколько отправлять статистику в output. Валидные значения единиц измерения - "ns", "us" (или "µs"), "ms", "s", "m", "h".|
|*counts*|перечисление _полей_, по которым надо строить счетчики по уникальным значениям|
|*aggregates*|перечисление _полей_, по которым будут собираться данные для групповых операций. Список доступных групповых операций описан ниже|
|*filters*|перечисление фильтров, по которым будут считаться метрики. Таким образом можно в отдельности считать метрики по каждому фильтру. Описание формата фильтра описано ниже|
|*output*|перечисление методов отправки результатов. У каждого отправителя  может быть своя настройка. Список доступных отправителей и способе их настройки описан ниже|


*input*

one of:

* file
* syslog
* stdin:nowait

**Filter**

|поле|описание|
|----|------|
|*filter*| регулярное выражение, описывающее, какие строки должны попасть под фильтр |
|*prefix*| префикс, который будет у ключа в output. |
|*items*| массив. перечисление метрик, которые надо посчитать и отправить в output |
|*items[].field*| названия поля. Соответствует полям из глобального регулярного выражения _regexp_ |
|*metrics*| перечисление метрик, которые надо посчитать для поля _field_|

**Output**

На данный момент доступно 2 отправщика: console и zabbix
console не имеет настроек, заббикс имеет следующие настройки
zabbix_host - хост сервера zabbix, 
zabbix_port - порт сервера zabbix, 
host - имя хоста, которым будет представляться приложение при отправке результатов

общий формат отправщика:

```
{
"type": "output_name",
"settings": {"output_config1":"output_config_value1"}
}
```

в случае отправщика console надо оставить объект settings пустым (settings:{})

**Формат ключа в отправщик**
{prefix}{field}{metric}

**Список доступных операций со счетчиками (counts):**

* cps_{val} - кол-во элементов по уникальному значению _{val}_ в секунду для поля _field_
* uniq - кол-во уникальных значений за период
* percentage_{val} - процент по уникальному _{val}_ за съем для поля _field_

**Список доступных групповых операци (aggregated):**

Сохраяняет все значения из поля (с плавающей запятой)
 и позволяет применить следующие операции:

* min - минимальное значение по полю
* max - максимальное значение по полю
* avg - среднее значение по полю
* ips (items per second)
* len (кол-во элементов в группе), 
* cent_{N} - посчитать N-ый перцентиль



Запуск
------

```
./access_logs_stats -c config.json
```


todo
-----------------

make english doc

make tests for sender 

make normal syslog parser

make conf.d/*.json for multiple instances of app

make normal exit after one tick
