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
scrape https://www.devports.de
```

will produce a directory scrape inside current folder and store scraped content there. Default waiting after a request is 250ms, to not to stress the web server. You can change this behavior.

## advanced usage

```zsh
scrape -r 500 -t 5 https://www.devports.de
```

will use 5 threads to scrape; after each request in a thread, it will wait 500ms before doing the next one. 