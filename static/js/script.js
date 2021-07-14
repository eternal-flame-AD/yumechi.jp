$('.dropdown-trigger').dropdown();

$(document).ready(function() {
    $('.sidenav').sidenav();
});

document.querySelectorAll("div.sensitive-img-unblur").forEach(el => {
    el.onclick = (e) => {
        el.parentElement.classList.add("acknowledged")
    }
})