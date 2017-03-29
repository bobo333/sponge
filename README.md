# Sponge

Amalgamates links from various internet sources.

Sources:
- New York Times
- Hacker News
- Reddit `r/golang`

## Installation
    go get github.com/bobo333/sponge

## Usage

    [GOPATH]/bin/sponge
    
Will write output to `/tmp/sponge.txt`

Required environment variables:
- `NYT_API_KEY` - New York Times (see [here](https://developer.nytimes.com/signup) to get one)
- `REDDIT_USERNAME` - Reddit (sign up with [reddit](https://www.reddit.com/) to get one)

*Note: If the above variables are not included, Sponge will run, but the associated sources will not be retrieved.*
