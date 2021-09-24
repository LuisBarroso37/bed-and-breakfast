function Prompt() {
  let toast = function (c) {
    const { msg = '', icon = 'success', position = 'top-end' } = c;

    const Toast = Swal.mixin({
      toast: true,
      title: msg,
      position: position,
      icon: icon,
      showConfirmButton: false,
      timer: 3000,
      timerProgressBar: true,
      didOpen: (toast) => {
        toast.addEventListener('mouseenter', Swal.stopTimer);
        toast.addEventListener('mouseleave', Swal.resumeTimer);
      },
    });

    Toast.fire({});
  };

  let success = function (c) {
    const { msg = '', title = '', footer = '' } = c;

    Swal.fire({
      icon: 'success',
      title: title,
      text: msg,
      footer: footer,
    });
  };

  let error = function (c) {
    const { msg = '', title = '', footer = '' } = c;

    Swal.fire({
      icon: 'error',
      title: title,
      text: msg,
      footer: footer,
    });
  };

  async function custom(c) {
    const { msg = '', title = '', icon = '', showConfirmButton = true } = c;

    const { value: formValues } = await Swal.fire({
      icon: icon,
      title: title,
      html: msg,
      backdrop: false,
      focusConfirm: false,
      showCancelButton: true,
      showConfirmButton: showConfirmButton,
      willOpen: () => {
        if (c.willOpen !== undefined) {
          c.willOpen();
        }
      },
      didOpen: () => {
        if (c.didOpen !== undefined) {
          c.didOpen();
        }
      },
    });

    if (formValues) {
      // If form was not closed by clicking on the 'Cancel' button
      if (formValues.dismiss === Swal.DismissReason.cancel) {
        c.callback(false);
      }

      if (formValues.value === '') {
        c.callback(false);
      }

      if (c.callback !== undefined) {
        c.callback(formValues);
      }
    }
  }

  return {
    toast: toast,
    success: success,
    error: error,
    custom: custom,
  };
}

function checkAvailabilityByRoom(roomId) {
  // Display modal to check availability for given room
  document
    .querySelector('#check-availability-button')
    .addEventListener('click', function () {
      let html = `
            <form id="check-availability-form" action="" method="post" novalidate class="needs-validation">
                <div class="row w-100" style="margin: 0;">
                    <div class="col">
                        <div class="row" id="reservation-dates-modal">
                            <div class="col">
                                <input disabled required class="form-control" type="text" name="start_date" id="start-date" placeholder="Arrival">
                            </div>
                            <div class="col">
                                <input disabled required class="form-control" type="text" name="end_date" id="end-date" placeholder="Departure">
                            </div>

                        </div>
                    </div>
                </div>
            </form>
            `;

      attention.custom({
        title: 'Choose your dates',
        msg: html,
        willOpen: () => {
          const elem = document.querySelector('#reservation-dates-modal');
          const rp = new DateRangePicker(elem, {
            format: 'yyyy-mm-dd',
            showOnFocus: true,
            minDate: new Date(),
          });
        },
        didOpen: () => {
          document.querySelector('#start-date').removeAttribute('disabled');
          document.querySelector('#end-date').removeAttribute('disabled');
        },
        callback: function (result) {
          // Retrieve form values and create object to be sent to the server
          let form = document.querySelector('#check-availability-form');
          let formData = new FormData(form);

          // Append CSRF token and room id
          formData.append('csrf_token', '{{.CsrfToken}}');
          formData.append('room_id', roomId);

          // Make API call
          fetch('/search-availability-json', {
            method: 'post',
            body: formData,
          })
            .then((res) => res.json())
            .then((data) => {
              if (data.ok) {
                attention.custom({
                  icon: 'success',
                  showConfirmButton: false,
                  msg:
                    '<p>Room is available</p>' +
                    '<p><a href="/book-room?id=' +
                    data.room_id +
                    '&start_date=' +
                    data.start_date +
                    '&end_date=' +
                    data.end_date +
                    '" class="btn btn-primary">' +
                    'Book now!</a></p>',
                });
              } else {
                attention.error({
                  msg: 'Room is not available',
                });
              }
            });
        },
      });
    });
}
