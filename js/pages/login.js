function join(){
    if ($('#nickname').val() != ""){
        location.href = location.protocol + "//" + location.host + "/join/" + $('#nickname').val();
    } else {
        $('#nickname').focus();
    }
}
