function runSource() {
    let source = document.getElementById('source');

    let formData = new FormData();
    formData.append('src', source.value);

    fetch("/execute/", {body: formData, method: "post", signal: AbortSignal.timeout(3000)})
        .then(function (response) {
            if (response.status !== 200) {
                window.location.reload();
                return;
            }
            return response.text();
        })
        .catch(function (error) {
            window.location.reload();
        })
        .then(function (html) {
            let result = document.getElementById('result');
            result.innerHTML = html;
        })
}
