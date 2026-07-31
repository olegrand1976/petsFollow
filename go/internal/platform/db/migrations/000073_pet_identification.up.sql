-- Optional pet identification: microchip + health book number + PDF (images→PDF pipeline).
ALTER TABLE pets.pets
  ADD COLUMN IF NOT EXISTS microchip_number TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS health_book_number TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS health_book_pdf_url TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS health_book_pdf_object_key TEXT NOT NULL DEFAULT '';
