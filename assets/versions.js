function callback(response) {
    const div = document.createElement('div');
    div.innerHTML = response.html;
    document.body.appendChild(div);
    const container = div.querySelector('.rst-versions');
    div.querySelector('.rst-current-version').addEventListener('click', function() {
        const classes = container.className.split(' ');
        const index = classes.indexOf('shift-up');
        if (index === -1) {
            classes.push('shift-up');
        } else {
            classes.splice(index, 1);
        }
        container.className = classes.join(' ');
    });
}

var script = document.createElement('script');
script.src = 'https://argocd-notifications.readthedocs.io/_/api/v2/footer_html/?'+
    'callback=callback&project=argocd-notifications&page=&theme=mkdocs&format=jsonp&docroot=docs&source_suffix=.md&version=' + (window['READTHEDOCS_DATA'] || { version: 'latest' }).version;

document.getElementsByTagName('head')[0].appendChild(script);
