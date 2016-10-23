$("input[name='file']").change(function() {
  if (currentlyUploading) return

  var file = this.files[0]

  if (file.size > 210000000) {
    // If larger than 200 MB
    displayError("Chosen file is too large", "The chosen file cannot be larger than 200 MB.")
    $("#file-field").addClass("error")
    $("button[type='submit']").addClass("disabled")
  } else {
    $("form").removeClass("error")
    $("#file-field").removeClass("error")
    $("button[type='submit']").removeClass("disabled")
  }
})

function submitForm(evt, fields) {
  evt.preventDefault()

  var file = $("input[name='file']").get(0).files[0]
  if (file && file.size > 210000000) {
    // If larger than 200 MB
    displayError("Chosen file is too large", "The chosen file cannot be larger than 200 MB.")
    $(".button-text").text("Try again")
    return
  }

  currentlyUploading = true
  $("button[type='submit']").addClass("disabled loading")
  $(".delete-button").addClass("disabled")
  $(".progress").show()

  var fd = new FormData()
  fd.append("title", fields.title)
  fd.append("author", fields.author)
  fd.append("subject", fields.subject)
  fd.append("update", fields.update)
  if (file) {
    fd.append("file", file)
  }

  uploadForm(fd, "/upload/note")
}

$("form").form({
    fields: {
      title: ["minLength[3]", 'regExp[^[^:/\\\\<>"\\|\\?\\*]+$]'],
      author: "empty",
      subject: "empty",
      file: "emptyUpdate"
    },
    onSuccess: submitForm
  }
)

$(".delete-button").click(function() {
  var noteID = $(".ui.modal").attr("note-id");
  $(".ui.modal").modal({
    onApprove: function() {
			$.post("/delete/note", {"id": noteID}, function(success) {
        window.location = "/notes"
      })
			return false;
		}
  }).modal("show");
})
