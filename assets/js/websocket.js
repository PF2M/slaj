var conn = new WebSocket('ws://'+ window.location.hostname +':8080/ws');
conn.onopen = function(e) {
    console.log("Connection established!");

    message = '{"type":"onPage", "content":"'+ window.location.pathname +'"}'

    conn.send(message);
};

conn.onmessage = function(e) {
	message = JSON.parse(e.data);
    $('.post-list').prepend(message.content);
    $('.post').first().hide().fadeIn(400);
};

$(document).on("pjax:end", function() {
    message = '{"type":"onPage", "content":"'+ window.location.pathname +'"}'

    conn.send(message);
})