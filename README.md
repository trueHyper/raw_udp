## raw_udp scanner
В файл `bin\default_modules.go` добавить импорт `_ "github.com/zmap/zgrab2/modules/raw_udp"`  
В modeules добавить директорию `raw_udp`

## Flags
| Флаг           | Значение                            |Пример    |
|----------------|-------------------------------------|----------|        
| `--set <set name>`        | Указывает номер набора протоколов   |--set set1
| `--fast`       | Быстрое сканирование                |--fast|
| `--custom`     | Указать список своих payload     |--custom="proto1 proto2"
| `--default`    | Указать первый для скана payload        |--default snmpv1
| `--use_port`   | Поиск по номеру порта               | --use_port
|`--payload-timeout`| Время на ожидание каждого ответа | --payload-timeout 2s

1. `--fast` врайтит все payload в соединение подряд и ждет первый ответ, записывает его. Можно использовать вместе с флагами `--set` `--custom` `--use_port` ❗**Если этот флаг включен, узнать от какого протокола сработал payload не получится**  

2. `--use_port` сканит только payload тех протоколов, у которых такой же порт. Если такого же порта не нашлось, он попробует найти похожий (*например не 161, а 162 - вдруг snmp просто немножко переехал*)

## probes.go
Есть мапа Set:
```go 
Sets = map[string][]string{
  "set1": {"snmpv1", "snmpv2", "snmpv3", "ntp", "dns"},
  "set2": {"snmpv1", "snmpv2", "snmpv3", "ntp", "dns", "netbios", "tftp", "upnp"},
  //...
}
```
В нее можно добавлять свои сеты и сканить по ним через флаг `--set`. Сам протокол и его payload должны быть обязательно добавлены в мапу `_payload_`:
```go
var _payload_ = map[string]Payload{
  "YOUR_PROTOCOL_NAME": {
    Port: 0,
    Data: []byte("YOUR PAYLOAD DATA"),
  },

  "snmpv1": {
    Port: 161,
    Data: []byte{
      0x30, 0x53, 0x02, 0x01, 0x00, 0x04, 0x06, 0x70, 0x75, 0x62, 0x6C, 0x69,
      0x63, 0xA0, 0x46, 0x02, 0x04, 0x00, 0x9E, 0xE5, 0x2F, 0x02, 0x01, 0x00,
      0x02, 0x01, 0x00, 0x30, 0x38, 0x30, 0x0C, 0x06, 0x08, 0x2B, 0x06, 0x01,
      0x02, 0x01, 0x01, 0x01, 0x00, 0x05, 0x00, 0x30, 0x0C, 0x06, 0x08, 0x2B,
      0x06, 0x01, 0x02, 0x01, 0x01, 0x05, 0x00, 0x05, 0x00, 0x30, 0x0C, 0x06,
      0x08, 0x2B, 0x06, 0x01, 0x02, 0x01, 0x01, 0x06, 0x00, 0x05, 0x00, 0x30,
      0x0C, 0x06, 0x08, 0x2B, 0x06, 0x01, 0x02, 0x01, 0x01, 0x04, 0x00, 0x05,
      0x00,
    },
  },
}
```

## Usage
```echo "42.3.184.132" | ./zgrab2 raw_udp --port 161 --payload-timeout 1s --custom="ntp quic snmpv1 snmpv2 snmpv3" --fast --default default  > result.json```

## Exmaple
<img width="1920" height="891" alt="изображение" src="https://github.com/user-attachments/assets/4e5f6112-1cfb-4100-90c4-ec95cc1c920c" />
