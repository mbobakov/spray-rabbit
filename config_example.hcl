// This file is example configuration for spray-rabbit util
elastic "es-na-index-sample" {
  hosts = [ "http://es-na:9200"]
  index = "sample-{{.Year}}.{{.Month}}.{{.Day}}"
  type = "mytype"
  batch_size = 1024
  batch_delay_ms = 1000
}

console "logInfo" {}

tcp "remoteTCP" {
  hosts = ["localhost:2120"]
  timeout_sec = 5
  only {
    exist = "message"
  }
}

rabbitmq {
  ack = false
  exchange = "myevents"
  uri = "amqp://user:password@mq-na:5672/userhost"
  key = "cool"
  prefetch_count = 512
  queue = "myevents"
  auto_delete = true
  spay_to = ["es-na-index-sample", "remoteTCP", "logInfo"]
}

rabbitmq {
  ack = false
  exchange = "myevents2"
  uri = "amqp://user2:password2@mq-na:5672/user2host"
  key = "coolest"
  prefetch_count = 512
  queue = "myevents2"
  auto_delete = true
  spay_to = ["logInfo"]
}
