const form = document.querySelector('#echo-form');
const responseOutput = document.querySelector('#response-output');

const formatJson = (value) => JSON.stringify(value, null, 2);

const renderResponse = (payload) => {
    responseOutput.textContent = formatJson(payload);
};

const renderError = (error) => {
    responseOutput.textContent = formatJson({ error });
};

form.addEventListener('submit', async (event) => {
    event.preventDefault();
    const data = new FormData(form);
    const message = data.get('message');

    if (!message) {
        renderError('Message is required.');
        return;
    }

    responseOutput.textContent = 'Sending…';

    try {
        const response = await fetch('/api/echo', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ message }),
        });

        if (!response.ok) {
            const errorText = await response.text();
            renderError(`Request failed: ${response.status} ${errorText}`);
            return;
        }

        const payload = await response.json();
        renderResponse(payload);
    } catch (error) {
        renderError(error.message || 'Unexpected error');
    }
});
