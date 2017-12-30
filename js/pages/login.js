function join(){
    input_nickname = $('#nickname').val()
    if(input_nickname.match(/[@#%&'"\(\)=\~\|`{}\[\]\*\+\.<>\\`\?;:\/]/)) {
        $("#join-alert").text("Symbols other than hyphens and underscores can not be used.");
        return;
    } else {
        $("#join-alert").text("");
    }
    if ($('#nickname').val() != ""){
        location.href = location.protocol + "//" + location.host + "/join/" + $('#nickname').val();
    } else {
        $('#nickname').focus();
    }
}
