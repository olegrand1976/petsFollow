ALTER TABLE pets.pets
  DROP COLUMN IF EXISTS health_book_pdf_object_key,
  DROP COLUMN IF EXISTS health_book_pdf_url,
  DROP COLUMN IF EXISTS health_book_number,
  DROP COLUMN IF EXISTS microchip_number;
