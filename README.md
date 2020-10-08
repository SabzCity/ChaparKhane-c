# Sabz City Platform
It is a platform to test some idea to create a society without physical or human base government!

## Code Rules
- Respect all [RFCs](https://github.com/SabzCity/RFCs) to write codes!
- Don't use any on-screen logs! 
- Don't exit program for any reason. just let back error to upper layers.
- Don't code in some way that need to exit program if some thing miss or unavailable! Instead error with 500 to user requests.

### Math Rules
- **Generally Accepted Accounting Principles** suggest 4 decimals precision!
```Math
PerCent		    >> Per Hundred			>> %
PerMille	    >> Per Thousand			>> ‰
PerMyriad	    >> Per Ten Thousand		>> ‱
PerCentMille    >> Per Hundred Thousand	>> pcm
PerMillion	    >> Per Million			>> ppm
```

## APIs

### Record Owner
Every where in platform see OwnerID, can be one of these user types:
- Person - Have 7 [intelligence](https://en.wikipedia.org/wiki/Intelligence) factors (IQ, EQ, ...) like Human, AI(robots,...) to talk, make decision, ...
- Organization - An organized group of people with a particular purpose, such as a business or government department.

### Authorization Models
We do authorization in method layer, So each method can have different authorization model, ever Public Methods!
So we do:
- Specify type of user send request: registered or guest
- Specify token type: owner or delegate
- Specify data type requested: User or Organization
- Specify some other data like RequestedUserLocation, ...

### User
- Check requested-user==owner-user
- Check delegate token
- Check user relations for friend, ... requested data

## GUI (Graphical User Interface)
We use some written GUI engine to compile GUI app to native devices OS and web! GUI engines just support HTML5 standard, So we develop this gui app by HTML5 architecture standards like progressive web app (PWA) with so improvement.

### Information architecture
#### Semantic content
We always care to write content in semantic way by all resources.
- https://html.spec.whatwg.org/multipage/
- https://www.w3.org/TR/rdfa-core/
- https://www.w3.org/TR/rdfa-lite/
- https://schema.org/
- https://search.google.com/structured-data/testing-tool

### Design Methodology
We respect semantic content and style content by design languages.
- https://developers.google.com/web/fundamentals/design-and-ux/responsive/

## VUI (Voice User Interface)
### UserAssistant service
UserAssistant can do a lot things with help of AI.
User can set a repeat action to do it automaitcly. e.g.
- Transfer money to another user in loop at specific time.
- Make invoices to renew service like domain, platform subscription, ...
- Make services to checkup monthly complex elevator, ...

## WWW

### URL Localize
We use hl(hreflang) parameter in url like others e.g. google, instagram, ...! Use [ISO 639](https://www.iso.org/iso-639-language-codes.html) for language codes and [ISO 3166](https://en.wikipedia.org/wiki/ISO_3166) for region codes.
- https://www.google.com/search?hl=en&q=sabz.city
- https://www.instagram.com/sabz.city/?hl=en

### URL standard
We use page name (service name) for indicate in URL. e.g   
` {domain}/{page-name}/{resource-uuid}?hl={lang}&{}={} `   
e.g.   
` www.sabz.city/product/123456789?name=product_name&hl=en `

### Serve
- Check productionInit() and devInit() comment state in main.go file for your purpose!
- Build by ```go build``` in a terminal
- Run build app by ```./sabz.city``` in a terminal
- Reload gui files by hit enter in the terminal in development version or send SIGUSR1(10) OS signal in both version anytime!

## Editors
### VSC
You must set this in settings to adjust formatter to fit this project before contributes
- html.format.wrapLineLength = 0

## [Bug Bounty Program](https://en.wikipedia.org/wiki/Bug_bounty_program)
If you find a bug please send a PR. But if you want to sell to others, first contact us by send email to ICTATSABZDOTCITY to negotiate about the price!
