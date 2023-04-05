# Notes

## What

Given a starting URL, the crawler should visit each URL it finds on the same domain. It should print each URL visited,
and a list of links found on that page.

### Constraints

- The crawler should be limited to one subdomain
    - (i.e, if the domain is monzo.com, monzo.com links should be followed)
    - (i.e, if the domain is monzo.com, community.monzo.com links should be printed but not followed)
    - (i.e, if the domain is monzo.com, facebook.com links should be printed but not followed)

### Considerations

Off the top of my head, these are things I may need to consider. I've put them in priority order in which I'll tackle
them. Some of them I won't implement if I run out of time.

- some or potentially all links will be relative
- duplicate links, either on the same page, or across pages (for example, nav links)
- rate limiting? I shouldn't hit sites too hard
- respect robots.txt
- dealing with pages that would have content rendered client-side, such as SPAs

#### Found during exploratory testing

- if we've already visited a link, we shouldn't visit it again if we see it with a trailing slash added
- not every file type needs to be visited (e.g. images, mp3s, etc)
- I decided to stream the results to stdout rather than printing them all at the end, because it can take a _long time_ to get to the end

## Testing approach

I could run tests against monzo.com because that was the example starting URL. Reasons not to:

- The content on any monzo.com page could easily change, making my tests brittle
- I'm starting with an approach that will have me implement rate limiting and respecting robots.txt later. I don't want to get blocked from any sites before then!

Instead, I'm going to start with testing against a fake http server. Once I've implemented rate limiting etc, I will
do some exploratory testing against monzo.com to check that it works, and also to find other edge cases I may have missed.

I'll start testing from the top down - so I'll start by writing an acceptance test for the CLI. I'm going to make the
output of the CLI JSON so that I can parse the results easily.
