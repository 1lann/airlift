$(".star-card").click(function() {
  var stars = parseInt(this.getAttribute("stars"));
  var starred = this.getAttribute("has-starred") == "true";
  var noteID = this.getAttribute("note-id");
  if (starred) {
    stars = stars - 1;
    starred = false;
    this.setAttribute("has-starred", "false");
    this.setAttribute("stars", stars);
    $(".star.icon").addClass("grey");
    $(".star.icon").removeClass("yellow");
  } else {
    stars = stars + 1;
    starred = true;
    this.setAttribute("has-starred", "true");
    this.setAttribute("stars", stars);
    $(".star.icon").removeClass("grey");
    $(".star.icon").addClass("yellow");
  }

  $.post("/notes/" + noteID + "/star", {starred: starred.toString()});

  $(".star-card .content").empty();
  if (stars == 1) {
    $(".star-card .content").append("1 star");
  } else {
    $(".star-card .content").append(stars + " stars");
  }

  if (starred) {
    $(".star-card .content").append($('<div class="sub header">You have starred this.</div>'));
  } else {
    $(".star-card .content").append($('<div class="sub header">Click here to star.</div>'));
  }
})
