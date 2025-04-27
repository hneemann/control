function runSource() {
    let source = document.getElementById('source');
    let result = document.getElementById('result');

    result.innerHTML = generateOutput(source.value);
    source.focus();
}
