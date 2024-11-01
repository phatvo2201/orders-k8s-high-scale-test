import http from 'k6/http';
import { sleep, check } from 'k6';

export let options = {
    vus: 1000, // Number of virtual users
    duration: '30s', // Test duration (adjust as needed)
};

function getRandomProductId() {
    return Math.floor(Math.random() * 1000) + 1; // Random ID between 1 and 1000
}

// Function to generate a random quantity
function getRandomQuantity() {
    return Math.floor(Math.random() * 3) + 1; // Random quantity between 1 and 3
}

// Function to generate random price as a float
function getRandomPrice() {
    return Math.random() * 100; // Random price between 0 and 100 as a float
}

export default function () {
    const productCount = Math.floor(Math.random() * 3) + 1; // Random number of products (1 to 3)
    const productIDs = [];
    const quantities = [];
    const prices = [];

    for (let i = 0; i < productCount; i++) {
        productIDs.push(getRandomProductId());
        quantities.push(getRandomQuantity());
        prices.push(getRandomPrice()); // Use the generated price directly as a float
    }

    const order = {
        id: "order_" + Math.floor(Math.random() * 1000), // Random order ID
        user_id: "user_" + Math.floor(Math.random() * 1000), // Example user ID
        product_ids: productIDs, // Use the generated product IDs slice
        quantities: quantities, // Use the generated quantities slice
        prices: prices, // Use the generated prices slice as floats
        status: "RECEIVED", // Changed to match the expected status
    };

    const res = http.post('http://localhost:8000/order', JSON.stringify(order), {
        headers: {
            'Content-Type': 'application/json',
        },
    });

    check(res, {
        'is status 200': (r) => r.status === 200,
    });

    console.log(`Response-----------------------: ${res.status}`);

    sleep(0.1);
}