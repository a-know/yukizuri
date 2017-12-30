$(function(){
    var socket = null;
    var exp = /(\b(https?|ftp|file):\/\/[-A-Z0-9+&@#\/%?=~_|!:,.;]*[-A-Z0-9+&@#\/%=~_|])/ig;
    var msgBox = $("#chatbox #message");
    var messages = $("#chat-box-messages");
    $("#chatbox").submit(function(){
        if (!msgBox.val()) return false;
        if (!socket) {
        alert("Error : Not established WebSocket connection.");
        return false;
        }
        socket.send(JSON.stringify({"Message": msgBox.val()}));
        msgBox.val("");
        return false;
    });
    if (!window["WebSocket"]) {
        alert("Error : This browser not supports WebSocket.");
    } else {
        socket = new WebSocket("wss://" + location.host + "/room");
        socket.onclose = function() {
            var icon = "./images/yukizuri-sys.png";
            var name = "Yukizuri-sys";
            messages.prepend(
                $("<div>").attr("class", "chat-box-msg left")
                .append(
                    $("<div>").attr("class", "chat-box-danger clearfix")
                    .append(
                        $("<span>").attr("class", "chat-box-name pull-left").text(name))
                    .append(
                        $("<span>").attr("class", "chat-box-timestamp pull-right").text((new Date()))
                    )
                )
                .append(
                    $("<img>").attr("class", "chat-box-img").attr("src", icon)
                )
                .append(
                    $("<div>").attr("class", "chat-box-text").text("Connection closed. Please reload this page.")
                )
            );
        }
        socket.onmessage = function(e) {
            var msg = JSON.parse(e.data);
            if (msg.Name == loginUserName) {
                messages.prepend(
                    $("<div>").attr("class", "chat-box-msg right")
                    .append(
                        $("<div>").attr("class", "chat-box-danger clearfix")
                        .append(
                            $("<span>").attr("class", "chat-box-name pull-right").text(msg.Name + " (" + msg.RemoteAddr + ")"))
                        .append(
                            $("<span>").attr("class", "chat-box-timestamp pull-left").text((new Date(msg.When)))
                        )
                    )
                    .append(
                        $("<img>").attr("class", "chat-box-img").attr("src", get_identicon(msg.Name))
                    )
                    .append(
                        $("<div>").attr("class", "chat-box-text").html(msg.Message.replace(/<("[^"]*"|'[^']*'|[^'">])*>/g,'').replace(exp,"<a href='$1'>$1</a>"))
                    )
                );
            } else {
                var icon = "";
                var name = "";
                if (msg.Name == "Yukizuri-sys") {
                    icon = "./images/yukizuri-sys.png";
                    name = "Yukizuri-sys";
                } else {
                    icon = get_identicon(msg.Name);
                    name = msg.Name + " (" + msg.RemoteAddr + ")";
                }
                messages.prepend(
                    $("<div>").attr("class", "chat-box-msg left")
                    .append(
                        $("<div>").attr("class", "chat-box-danger clearfix")
                        .append(
                            $("<span>").attr("class", "chat-box-name pull-left").text(name))
                        .append(
                            $("<span>").attr("class", "chat-box-timestamp pull-right").text((new Date(msg.When)))
                        )
                    )
                    .append(
                        $("<img>").attr("class", "chat-box-img").attr("src", icon)
                    )
                    .append(
                        $("<div>").attr("class", "chat-box-text").html(msg.Message.replace(/<("[^"]*"|'[^']*'|[^'">])*>/g,'').replace(exp,"<a href='$1'>$1</a>"))
                    )
                );
                // refresh members list
                $("#members-count").text(msg.CurrentMembers.length);
                $("#members-box").empty();
                for(let i = 0; i < msg.CurrentMembers.length; i++) {
                    var memberName = "";
                    if (loginUserName == msg.CurrentMembers[i]["name"]) {
                        memberName = msg.CurrentMembers[i]["name"] + " (You)";
                    } else {
                        memberName = msg.CurrentMembers[i]["name"];
                    }
                    $("#members-box")
                    .append(
                        $("<li>")
                        .append(
                            $("<img>").attr("src", get_identicon(msg.CurrentMembers[i]["name"]))
                        )
                        .append(
                            $("<span>").attr("class", "users-list-date").text(memberName)
                        )
                        .append(
                            $("<span>").attr("class", "users-list-date").text(msg.CurrentMembers[i]["remote_addr"])
                        )
                    );
                }
            }
        }
    }
});

function get_identicon(text) {
    var salt = 0;
    var rounds = 1;
    var size = 210;
    var outputType = "HEX";
    var hashtype = "SHA-512";
    var shaObj = new jsSHA(text+salt, "TEXT");
    var hash = shaObj.getHash(hashtype, outputType,rounds);
    var data = new Identicon(hash, size).toString();
    return 'data:image/png;base64,' + data;
}