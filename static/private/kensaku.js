window.onerror = e =>
    document.querySelector("#errmsg").innerText = e.message || e

function decodeQueryParam(val) {
    return decodeURIComponent(val.replace(/&/g, "%26").replace(/=/g, "%3D").replace(/\+/g, " "));
}

function stripPrefix(url, prefix) {
    if (url.startsWith(prefix))
        return url.slice(prefix.length)
    return url
}

function redirectToSearch(target, keyword) {
    console.log("query redirect",
        "target=" + target, "keyword=" + keyword)
    switch (target) {
        // GENERAL
        case "ddg":
        case "duckduckgo":
            window.location = "https://html.duckduckgo.com/?q=" + encodeURIComponent(keyword)
            break
        case "g":
        case "google":
            window.location = "https://www.google.com/search?q=" + encodeURIComponent(keyword)
            break
        case "gimg":
        case "gimage":
            window.location = "https://www.google.com/search?tbm=isch&q=" + encodeURIComponent(keyword)
            break
        case "yt":
        case "youtube":
            window.location = "https://www.youtube.com/results?search_query=" + encodeURIComponent(keyword)
            break
        case "mdn":
            window.location = "https://developer.mozilla.org/en-US/search?q=" + encodeURIComponent(keyword)
            break
        // JITEN
        case "jisho":
            window.location = "https://jisho.org/search/" + encodeURIComponent(keyword)
            break
        case "daijisen":
        case "djs":
        case "gokugo":
        case "goojp":
            window.location = "https://dictionary.goo.ne.jp/srch/jn/" + encodeURIComponent(keyword) + "/m0u/"
            break
        case "waei":
        case "eiwa":
        case "gooen":
            window.location = "https://dictionary.goo.ne.jp/srch/en/" + encodeURIComponent(keyword) + "/m0u/"
            break
        // ACADEMICS
        case "utlib":
            window.location = "https://search.lib.utexas.edu/discovery/search?vid=01UTAU_INST:SEARCH&tab=Everything&search_scope=MyInst_and_CI&offset=0&radios=resources&query=any,contains," + encodeURIComponent(keyword)
            break
        case "gs":
        case "gscolar":
            document.location = "https://scholar.google.com/scholar?q=" + encodeURIComponent(keyword)
            break
        case "pdbj":
            if (keyword.match(/^[a-zA-Z0-9]{4}$/))
                window.location = "https://pdbj.org/mine/summary/" + encodeURIComponent(keyword)
            else if (keyword.match(/^[a-zA-Z]{3}$/))
                window.location = "https://pdbj.org/chemie/summary/" + encodeURIComponent(keyword)
            else
                window.location = "https://pdbj.org/search/pdb?query=" + encodeURIComponent(keyword)
            break
        case "pdb":
        case "rcsb":
            if (keyword.match(/^[a-zA-Z0-9]{4}$/))
                window.location = "https://www.rcsb.org/structure/" + encodeURIComponent(keyword)
            else if (keyword.match(/^[a-zA-Z]{3}$/))
                window.location = "https://www.rcsb.org/ligand/" + encodeURIComponent(keyword)
            else
                window.location = "https://www.rcsb.org/search?request=" + encodeURIComponent(JSON.stringify({
                    "query": {
                        "type": "terminal",
                        "label": "full_text",
                        "service": "full_text",
                        "parameters": {
                            "value": keyword
                        }
                    },
                    "return_type": "entry",
                    "request_options": {
                        "pager": {
                            "start": 0,
                            "rows": 25
                        },
                        "scoring_strategy": "combined",
                        "sort": [{
                            "sort_by": "score",
                            "direction": "desc"
                        }]
                    }
                }))
            break

        default:
            redirectToSearch("ddg", target + (keyword ? " " + keyword : ""))
    }
}

function handleQuery(query) {
    query = query.replaceAll("　", " ").replace(/\s*$/, "").replace(/^\s*/, "")
    let queryWords = query.split(" ")
    let target = queryWords[0]
    if (queryWords[0].startsWith("_")) {
        target = "ddg"
        queryWords[0] = queryWords[0].slice(1)
    } else
        queryWords = queryWords.slice(1)
    if (!redirectToSearch(target, queryWords.join(" ")))
        document.querySelector("#search-input").value = query
}

function queryFromURLHash() {
    const queryFromHash = decodeQueryParam(stripPrefix(new URL(document.URL).hash, "#"))
    if (queryFromHash)
        handleQuery(queryFromHash)
}
window.addEventListener('hashchange', queryFromURLHash)
queryFromURLHash()


document.querySelector("#query-form").onsubmit = e => {
    e.preventDefault()
    handleQuery(document.querySelector("#search-input").value)
}