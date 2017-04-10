# Sponge

Amalgamates links from various internet sources.

Sources:
- New York Times
- Washington Post
- The Wall Street Journal
- The Economist
- Techcrunch
- Hacker News
- Reddit
    - `r/golang`
    - `r/python`
    - `r/sysadmin`
    - `r/programming`
    - `r/liverpoolfc`

*Note: Sources from News API and Reddit are trivially configurable.*


## Installation
    go get github.com/bobo333/sponge

## Usage

    [GOPATH]/bin/sponge
    
Will write output to `[default temp directory]/sponge.html` by default.

Required environment variables:
- `NEWS_API_KEY` - News API (see [here](https://newsapi.org/) to get one)
- `REDDIT_USERNAME` - Reddit (sign up with [reddit](https://www.reddit.com/) to get one)

*Note: If the above variables are not included, Sponge will run, but the associated sources will not be retrieved.*


### Options

See `[GOPATH]/bin/sponge --help` for details.

- Supports emailing output to address instead of writing to file. Requires additional environment vars:
    - `MAILGUN_API_KEY`
    - `MAILGUN_DOMAIN`
- Supports text-formatted output instead of html.
- Includes configurable number of items to fetch per source.
