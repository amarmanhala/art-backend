# Production Deployment Checklist

## Stripe
- Set `STRIPE_SECRET_KEY` to the **live** secret key from Stripe Dashboard.
- Set `STRIPE_WEBHOOK_SECRET` from the **live** webhook endpoint, not the local listener.
- Update `STRIPE_SUCCESS_URL` to the production frontend URL.
- Update `STRIPE_CANCEL_URL` to the production frontend URL.
- Keep `STRIPE_ALLOWED_SHIPPING_COUNTRIES` aligned with the markets you actually support.
- Register the production webhook endpoint in Stripe and listen for:
  - `checkout.session.completed`
  - `checkout.session.expired`

## Frontend URLs
- Replace all local callback URLs with production URLs.
- Make sure the success page exists in the deployed frontend.
- Make sure the cancel/cart page exists in the deployed frontend.

## Backend Environment
- Confirm these are set in the production runtime:
  - `APP_PORT`
  - `CORS_ALLOWED_ORIGINS`
  - `DB_HOST`
  - `DB_PORT`
  - `DB_USER`
  - `DB_PASSWORD`
  - `DB_NAME`
  - `DB_SSLMODE`
  - `AZURE_STORAGE_ACCOUNT_NAME`
  - `AZURE_STORAGE_ACCOUNT_KEY`
  - `AZURE_STORAGE_CONTAINER`
  - `AZURE_PRODUCT_IMAGES_CONTAINER`
  - `AZURE_FRAME_IMAGES_CONTAINER`
  - `AZURE_ART_STYLES_CONTAINER`
  - `STRIPE_SECRET_KEY`
  - `STRIPE_WEBHOOK_SECRET`
  - `STRIPE_SUCCESS_URL`
  - `STRIPE_CANCEL_URL`

## Azure Blob Storage
- Confirm production containers exist:
  - `carousel-images`
  - `product-images`
  - `frame-images`
  - `art-styles`
- Confirm public URLs point to the production storage account.
- Confirm upload SAS generation still matches the live account name/key.

## Database
- Run all migrations on the production database before deploying the app.
- Verify these tables exist:
  - `products`
  - `product_images`
  - `product_variants`
  - `carts`
  - `cart_items`
  - `orders`
  - `order_items`
  - `frames`
  - `frame_images`
  - `art_styles`
  - `contact_requests`
- Verify Stripe checkout can create and finalize orders against production data.

## Runtime Checks
- Recreate or restart the app container after env changes.
- Verify checkout creates an order before redirecting to Stripe.
- Verify the webhook marks the order paid and clears the cart.
- Verify product create/edit still works with variants.
- Verify cart add uses `product_variant_id`.

## Cleanup Before Go-Live
- Remove any test-only webhook listener usage from local notes/scripts.
- Remove any hard-coded localhost URLs from frontend or backend configs.
- Confirm test cards and sandbox references are not used in production docs.
