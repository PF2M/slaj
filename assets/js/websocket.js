var conn = new WebSocket('ws://'+ window.location.hostname +':8080/ws');
conn.onopen = function(e) {
    console.log("Connection established!");

    message = '{"type":"onPage", "content":"'+ window.location.pathname +'"}'

    conn.send(message);
};

conn.onmessage = function(e) {
	message = JSON.parse(e.data);
	if (message.type == "comment") {
		$('.reply-list').append(message.content);
		$('.post').last().hide().fadeIn(400);
		commentCount = parseInt($('.reply-count').text());
		$('.reply-count').text(commentCount+1);
	} else if (message.type == "commentPreview") {
		$('#'+ message.id).find('.recent-reply-content').remove();
		$('#'+ message.id).find('.post-meta').after(message.content);
		commentCount = parseInt($('#'+ message.id).find('.reply-count').text());
		$('#'+ message.id).find('.reply-count').text(commentCount+1);
		if (commentCount > 1) {
			$('#'+ message.id).find('.recent-reply-content').prepend('<div class="recent-reply-read-more-container" tabindex="0">View all comments ('+ (commentCount+1) +')</div>')
		}
	} else if (message.type == "post") {
		$('.post-list').prepend(message.content);
		$('.post').first().hide().fadeIn(400);
	}
};

$(document).on("pjax:end", function() {
    message = '{"type":"onPage", "content":"'+ window.location.pathname +'"}'

    conn.send(message);
})