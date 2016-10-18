$(".complete-card").click(function() {
  var nowCompleted = $(this).hasClass("orange");

  if (nowCompleted) {
    $(this).removeClass("orange");
    $(this).addClass("green");
    $(".complete-card .orange").removeClass("orange").addClass("green");
    $(".complete-card .icon").removeClass("remove").addClass("checkmark");
    $(".complete-card .content").empty().append("Practice paper completed");
  } else {
    $(this).removeClass("green");
    $(this).addClass("orange");
    $(".complete-card .green").removeClass("green").addClass("orange");
    $(".complete-card .icon").removeClass("checkmark").addClass("remove");
    $(".complete-card .content").empty().append("Practice paper not completed");
    $(".complete-card .content").append(
        $('<div class="sub header">Click here when you complete this paper.</div>'));
  }

  var paperID = this.getAttribute("paper-id");
  $.post("/papers/" + paperID + "/complete", {completed: nowCompleted.toString()});
})
