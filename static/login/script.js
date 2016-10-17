var focused = false;

$("input[type='password']").blur(function() {
	focused = false;
	$(".floating-above").css("opacity", "0");
	setTimeout(function() {
		if (!focused) {
			$(".floating-above").css("z-index", "-100");
		}
	}, 400);
});

$("input[type='password']").focus(function() {
	focused = true;
	$(".floating-above").css("z-index", "100");
	$(".floating-above").css("opacity", "1");
});

$("form").on("submit", function() {
	$("button").addClass("loading disabled");
});
