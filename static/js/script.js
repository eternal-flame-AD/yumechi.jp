$('.dropdown-trigger').dropdown();

$(document).ready(function() {
    $('.sidenav').sidenav();
});

document.querySelectorAll("div.sensitive-img-unblur").forEach(el => {
    el.onclick = (e) => {
        el.parentElement.classList.add("acknowledged")
    }
})

function toggleTategumi(selector) {
    selector = selector || "div.content"
    let existingClass = document.querySelector(selector).classList;
    if (existingClass.contains("tategumi")) {
        existingClass.remove("tategumi");
        return false;
    }
    existingClass.add("tategumi");
    return true;
}