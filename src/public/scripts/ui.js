function set_btn_processing(spinner, is_processing) {
  if (is_processing) {
    spinner.style.opacity = 1;
    return;
  }
  spinner.style.opacity = 0;
}

export default { set_btn_processing };
