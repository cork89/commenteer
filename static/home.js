

function validateLink() {
    let url = document.getElementById("link");
    
    let _, error = isValidRedditUrl(url.value)
    if (error) {
        document.getElementById("linkError").innerText = error;
    } else {
        let [subreddit, article, comment] = getLinkTokens(url.value)
        window.location.href = `/r/${subreddit}-${article}-${comment}/submit`
    }
}

function isValidRedditUrl(input) {
    let url;
    try {
        url = new URL(input);
        if (url.host !== "reddit.com" && url.host !== "www.reddit.com") {
            return false, "Not a reddit link";
        }
    } catch (_) {
        return false, "Not a valid url";
    }
    return url.protocol === "http:" || url.protocol === "https:", "";
}

function getLinkTokens(link) {
    let tokens = link.split("/")
    let [https, url, r, subreddit, comments, article, title, comment] = tokens.filter(item => item != '')
    return [subreddit, article, comment]
}