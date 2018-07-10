function checkForm() {
    if($("textarea[name=body]").val().length > 0) {
        Olv.Form.toggleDisabled($("input.post-button"), false);
    } else {
        Olv.Form.toggleDisabled($("input.post-button"), true);
    }
}

function error(error) {
    $("input[name=image]").val("");
    $(".file-button").val(null);
    $(".preview-container").css("display", "none");
    checkForm();
    $(".file-button").removeAttr("disabled");
    $(".file-upload-button").text("Upload");
    Olv.showMessage("Error", error);
}

$(".file-button").on("change", function() {
    if(this.files.length) {
        Olv.Form.toggleDisabled($("input.post-button"), true);
        $(".file-button").attr("disabled", "disabled");
        $(".file-upload-button").text("Uploading...");
        var reader = new FileReader();
        reader.readAsDataURL(this.files[0]);
        reader.onload = function () {
            $.post("/upload", reader.result, function(data) {
                $("input[name=image]").val(data);
                $(".preview-container img").attr("src", data)
                $(".preview-container").attr("style", "");
                checkForm();
                $(".file-button").removeAttr("disabled");
                $(".file-upload-button").text("Upload");
            }, "text").fail(function(error) {
                error(error);
            });
        };
        reader.onerror = function (error) {
            error(error);
        };
    } else {
        $("input[name=image]").val("");
        $(".preview-container").css("display", "none");
        checkForm();
    }
});