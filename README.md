# about

scrape was written to scrape a site fast and also store header information to can grep them. It is part of an ethical hacking reconnaissance part. So sure, you can scrape the whole domain with each path that could be found, but you can also scrape only some specific paths dependent to your own experienced template.

# install

At the moment, there is no real install script provided. You need to have go installed (locally I have developed and tested it with version 1.14.3), so you can call

```zsh
go install
```

After the binary was produced, you can make an alias for your terminal, if you want to - I made an alias called "scrape" and following examples will use this alias.

# usage

## list options

```zsh
scrape
```


## basic usage

``` zsh
scrape -deep https://www.devports.de
```

will produce a directory scrape inside current folder and store scraped content there. Default waiting after a request is 250ms, to not to stress the web server. You can change this behavior.

## advanced usage

### Use more threads:

```zsh
scrape -deep -r 500 -t 5 https://www.devports.de
```

will use 5 threads to scrape; after each request in a thread, it will wait 500ms before doing the next one. 

### scrape a specif path template

For example, you want to have a lookup for specific backend paths, you can do it with following command:

```zsh
scrape -tp ./template https://www.devports.de
```

Inside the ./template file, you have to  provide a file with some paths; each one is a new line. Something like this for example:

```
/robots.txt
/backend/login
/login
```