
# spray-rabbit
'''spray-rabbit''' is tool for sending JSON-messages from Rabbit-MQ to different destinations like TCP-socket, ElasticSearch and etc.

### Install
``` go get https://github.com/mbobakov/spray-rabbit```

or download via releases page

### Implemented outputs(Sprays)
  - [Console](https://github.com/mbobakov/spray-rabbit/console-spray/). Print message as Info-log event
  - [TCP](https://github.com/mbobakov/spray-rabbit/tcp-spray/). Send messages to TCP-socket. Messages delimited by '\n'
  - [ElasticSearch](https://github.com/mbobakov/spray-rabbit/elastic-spray/). Send messages via bulk-API

### Configuration
For now all configuration is placed in one file. File must be in [HCL-syntax](https://github.com/hashicorp/hcl)

[Example config](https://github.com/mbobakov/spray-rabbit/config_example.hcl)

Format of configuration for the input:

    rabbitmq {
        uri             <string>        AMQP connection URL like "amqp://user:password@mq-na:5672/userhost"
        exchange        <string>        Rabbitmq Exchange for binding fetching queue
        queue           <string>        Rabbitmq Queue for fetching messages
        key             <string>        Rabbitmq Routing key
        ack             <bool>          Shoud we ack fetch for Rabbitmq
        prefetch_count  <int>           Prefetch count
        auto_delete     <bool>          Delete the Rabbitmq queue if no consumers on it
        spay_to         <[] of string>  Send messages via this outputs
    }

### TODO
    - Tests
    - More clear documentation

### Contibuting
Go >= 1.7 is requirement. This tool use [dep](https://github.com/golang/dep) for dependency management. Pull Request are welcome. ;) 
