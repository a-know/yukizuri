function join(){
    if ($('#nickname').val() != ""){
        location.href = location.protocol + "//" + location.host + "/auth/join/" + $('#nickname').val();
    } else {
        $('#nickname').focus();
    }
}
