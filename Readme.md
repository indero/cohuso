# Cohuso

Cohuso is a small go application inspired by an old [PHP application](http://geekbox.ch/voip-telefon-snom-370/) that can be used to reverse lookup swiss phone numbers via [tel.search.ch api](https://tel.search.ch/api/help.html#response)and display them on snom phones.

## Installation

## Usage
To try it out:
```bash
go run server.go
```
Run it with an [API Key](https://tel.search.ch/api/getkey) (recommended):
```bash
go run server.go --api-key "2aa83dd1bba136430dc4f75e7715c577e4caa494"
```

To change the listen address (default: :8080):
```
go run server.go --http-listen-addr=127.0.0.1:8443
```

### Make a request
```bash
$ curl localhost:8080/call.php\?caller=0581346620
```
The response:
```xml
<SnomIPPhoneText>
	<Title>Anruf von 0581346620</Title>
	<Prompt>Anruf von 0581346620</Prompt>
	<Text>Schweizer Radio und Fernsehen (SRF)&#xA;    Regionalstudio Zentralschweiz&#xA;    Inseliquai 8&#xA;    6005 Luzern LU&#xA;    *058 134 66 20</Text>
</SnomIPPhoneText>
```

## Docker

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT](https://choosealicense.com/licenses/mit/)