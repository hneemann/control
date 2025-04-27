function runSource() {
    let source = document.getElementById('source');
    let result = document.getElementById('result');

    fetchHelper("/execute/", source.value, a => {
        result.innerHTML = a;
        source.focus();
    });
}
