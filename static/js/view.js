let loadMainContainer = () => {
    return fetch('/userinfo', {credentials: 'include'})
        .then((response) => response.json())
        .then((response) => {
            if(response.status === 'ok') {
                $('#name').html(response.name)
                $('#username').html(response.username)
                $('#registerContainer').hide();
                $('#loginContainer').hide();
                $('#mainContainer').show();
            } else {
                alert(`Error! ${response.message}`)
            }
        })
}

let checkIfLoggedIn = () => {
    return fetch('/userinfo', {credentials: 'include'})
        .then((response) => response.json())
        .then((response) => {
            if(response.status === 'ok') {
                return true
            } else {
                return false
            }
        })
}

$('#logoutButton').click(() => {
    fetch('/logout', {credentials: 'include'});
    document.cookie = 'session_id=; Max-Age=-99999999;';  
    $('#mainContainer').hide();
    $('#loginContainer').show();
})
