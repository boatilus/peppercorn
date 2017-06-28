document.addEventListener('DOMContentLoaded', function() {
  let MFADisable = document.getElementById('mfa_disable');
  if (MFADisable !== null) {
    MFADisable.addEventListener('click', MFADisableClickHandler);
  }
});

let MFADisableClickHandler = function(event) {
  event.preventDefault();

  if (window.confirm("Are you sure you want to disable two-factor authentication?")) {
    document.getElementById('mfa_duration_form').submit();
  }
}