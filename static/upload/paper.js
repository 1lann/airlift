function checkSizes() {
  $("form").removeClass("error")
  $("#file-field").removeClass("error")
  $("button[type='submit']").removeClass("disabled")

  var paper = $("#questions-field input").get(0).files[0]
  if (paper && paper.size > 210000000) {
    displayError("Chosen file is too large", "The questions paper cannot be larger than 200 MB.")
    $("#file-field").addClass("error")
    $("button[type='submit']").addClass("disabled")
    return
  }

  var source = $("#source-field input").get(0).files[0]
  if (source && source.size > 210000000) {
    displayError("Chosen file is too large", "The source booklet cannot be larger than 200 MB.")
    $("#file-field").addClass("error")
    $("button[type='submit']").addClass("disabled")
    return
  }

  var solutions = $("#solutions-field input").get(0).files[0]
  if (solutions && solutions.size > 210000000) {
    displayError("Chosen file is too large", "The paper's solutions cannot be larger than 200 MB.")
    $("#file-field").addClass("error")
    $("button[type='submit']").addClass("disabled")
    return
  }
}

$("input[type='file']").change(checkSizes)

function submitForm(evt, fields) {
  evt.preventDefault()

  var questionsFile = $("#questions-field input").get(0).files[0]
  if (questionsFile && questionsFile.size > 210000000) {
    // If larger than 200 MB
    displayError("Chosen file is too large", "The questions paper cannot be larger than 200 MB.")
    $(".button-text").text("Try again")
    return
  }

  var sourceFile = $("#source-field input").get(0).files[0]
  if (sourceFile && sourceFile.size > 210000000) {
    displayError("Chosen file is too large", "The source booklet cannot be larger than 200 MB.")
    $("#file-field").addClass("error")
    $("button[type='submit']").addClass("disabled")
    return
  }

  var solutionsFile = $("#solutions-field input").get(0).files[0]
  if (solutionsFile && file.size > 210000000) {
    // If larger than 200 MB
    displayError("Chosen file is too large", "The paper's solutions cannot be larger than 200 MB.")
    $(".button-text").text("Try again")
    return
  }

  currentlyUploading = true
  $("button[type='submit']").addClass("disabled loading")
  $(".progress").show()

  var fd = new FormData()

  fd.append("title", fields.title)
  fd.append("author", fields.author)
  fd.append("subject", fields.subject)
  fd.append("year", fields.year)
  fd.append("update", fields.update)
  if (sourceFile) {
    fd.append("source", sourceFile)
  }
  if (solutionsFile) {
    fd.append("solutions", solutionsFile)
  }
  if (questionsFile) {
    fd.append("questions", questionsFile)
  }

  uploadForm(fd, "/upload/paper")
}

$("form").form({
    fields: {
      title: ["minLength[3]", 'regExp[^[^:/\\\\<>"\\|\\?\\*]+$]'],
      author: "empty",
      subject: "empty",
      year: "integer[1990..2016]",
      questions: "emptyUpdate"
    },
    onSuccess: submitForm
  }
)
