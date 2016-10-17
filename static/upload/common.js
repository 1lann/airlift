$("select.dropdown").dropdown()

var progressBar = $(".ui.progress")
var progressText = $(".progress .label")
var currentlyUploading = false

function setProgress(color, text, active) {
  progressBar.removeClass("red")
  progressBar.removeClass("blue")
  progressBar.removeClass("teal")
  progressBar.removeClass("green")
  progressBar.addClass(color)

  if (color == "red") {
    $(".progress.percentage").text("")
  }

  progressText.text(text)

  if (active) {
    progressBar.addClass("active")
  } else {
    progressBar.removeClass("active")
  }
}

function displayError(title, message) {
  $("form").addClass("error")
  var header = $('<div class="header"></div>')
  header.text(title)
  $(".error.message").empty().append(header).append(message)
}

function uploadForm(fd, url) {
  $.ajax({
    xhr: function() {
      var xhr = new window.XMLHttpRequest();
      xhr.upload.addEventListener("progress", function(evt) {
        if (evt.lengthComputable) {
          var percentComplete = Math.round((evt.loaded / evt.total) * 100);
          $(".progress .bar").css("width", percentComplete + "%")
          $(".progress.percentage").text(percentComplete + "%")
          setProgress("blue", "Uploading...", true)

          if (percentComplete === 100) {
            setProgress("teal", "Almost there, processing...", true)
          }
        }
      }, false);

      return xhr;
    },
    url: url,
    type: "POST",
    data: fd,
    contentType: false,
    processData: false,
    cache: false,
    success: function(result) {
      currentlyUploading = false
      setProgress("green", "Upload complete!", false)

      if (url == "/upload/note") {
        window.location = "/notes/" + result.id
      } else if (url == "/upload/paper") {
        window.location = "/papers/" + result.id
      } else {
        console.log("airlift: unknown upload type")
      }
    },
    error: function(result) {
      currentlyUploading = false
      var errorMessage = "Unknown error. Contact Chuie for help."

      if (result.status == 500) {
        errorMessage = "Server error, contact Chuie for help."
      } else if (result.status == 413) {
        errorMessage = "The file is too large, it cannot be more than 200 MB."
      } else if (result.status == 415) {
        errorMessage = "The file is not a valid PDF."
      } else if (result.status == 406) {
        errorMessage = "Your form has errors in it. Did you bypass form validation?"
      }

      setProgress("red", "Upload failed!", false)
      displayError("Upload failed", errorMessage)

      $("button[type='submit']").removeClass("disabled loading")
      $(".button-text").text("Try again")

      console.log("upload failed", result)
    }
  });
}


$("select.dropdown").dropdown()

var progressBar = $(".ui.progress")
var progressText = $(".progress .label")
var currentlyUploading = false

function setProgress(color, text, active) {
  progressBar.removeClass("red")
  progressBar.removeClass("blue")
  progressBar.removeClass("teal")
  progressBar.removeClass("green")
  progressBar.addClass(color)

  if (color == "red") {
    $(".progress.percentage").text("")
  }

  progressText.text(text)

  if (active) {
    progressBar.addClass("active")
  } else {
    progressBar.removeClass("active")
  }
}

function displayError(title, message) {
  $("form").addClass("error")
  var header = $('<div class="header"></div>')
  header.text(title)
  $(".error.message").empty().append(header).append(message)
}

$.fn.form.settings.rules.emptyUpdate = function(value) {
  if ($("input[name='update']").val()) {
    return true
  }

  if (value) {
    return true
  }

  return false
};
