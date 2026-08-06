BEGIN;

-- ---------------------------------------------------------
-- Companies
-- ---------------------------------------------------------

INSERT INTO companies (
    id,
    name,
    website,
    industry,
    location,
    notes
)
VALUES
    (
        '11111111-1111-1111-1111-111111111111',
        'Spotify',
        'https://www.spotify.com',
        'Music Technology',
        'Stockholm, Sweden',
        'Interesting company for backend and platform engineering internships.'
    ),
    (
        '22222222-2222-2222-2222-222222222222',
        'Volvo Cars',
        'https://www.volvocars.com',
        'Automotive Technology',
        'Gothenburg, Sweden',
        'Potential opportunities involving connected vehicles and cloud services.'
    ),
    (
        '33333333-3333-3333-3333-333333333333',
        'Klarna',
        'https://www.klarna.com',
        'Fintech',
        'Stockholm, Sweden',
        'Relevant for backend, payments, and distributed systems experience.'
    ),
    (
        '44444444-4444-4444-4444-444444444444',
        'Ericsson',
        'https://www.ericsson.com',
        'Telecommunications',
        'Stockholm, Sweden',
        'Possible internships in networking, cloud infrastructure, and backend development.'
    ),
    (
        '55555555-5555-5555-5555-555555555555',
        'Northvolt',
        'https://northvolt.com',
        'Battery Technology',
        'Stockholm, Sweden',
        'Interesting engineering company with software and data-related roles.'
    )
ON CONFLICT (name) DO UPDATE
SET
    website = EXCLUDED.website,
    industry = EXCLUDED.industry,
    location = EXCLUDED.location,
    notes = EXCLUDED.notes,
    updated_at = NOW();

-- ---------------------------------------------------------
-- Applications
-- ---------------------------------------------------------

INSERT INTO applications (
    company_id,
    role,
    status,
    applied_at
)
SELECT
    c.id,
    seed.role,
    seed.status,
    seed.applied_at
FROM (
    VALUES
        (
            'Spotify',
            'Backend Engineering Intern',
            'applied',
            DATE '2026-07-10'
        ),
        (
            'Volvo Cars',
            'Software Developer Intern',
            'interview',
            DATE '2026-06-25'
        ),
        (
            'Klarna',
            'Go Backend Intern',
            'rejected',
            DATE '2026-06-12'
        ),
        (
            'Ericsson',
            'Cloud Software Intern',
            'offer',
            DATE '2026-07-01'
        ),
        (
            'Northvolt',
            'Platform Engineering Intern',
            'draft',
            NULL
        )
) AS seed(company_name, role, status, applied_at)
JOIN companies c
    ON c.name = seed.company_name
WHERE NOT EXISTS (
    SELECT 1
    FROM applications a
    WHERE a.company_id = c.id
      AND a.role = seed.role
);

-- ---------------------------------------------------------
-- Contacts
-- ---------------------------------------------------------

INSERT INTO contacts (
    company_id,
    name,
    email,
    linkedin_url,
    role,
    notes
)
SELECT
    c.id,
    seed.contact_name,
    seed.email,
    seed.linkedin_url,
    seed.contact_role,
    seed.notes
FROM (
    VALUES
        (
            'Spotify',
            'Anna Lindberg',
            'anna.lindberg@example.com',
            'https://www.linkedin.com/in/anna-lindberg-example',
            'Technical Recruiter',
            'Met at a university career fair.'
        ),
        (
            'Volvo Cars',
            'Johan Berg',
            'johan.berg@example.com',
            'https://www.linkedin.com/in/johan-berg-example',
            'Engineering Manager',
            'Contacted after submitting the internship application.'
        ),
        (
            'Klarna',
            'Sara Nilsson',
            'sara.nilsson@example.com',
            'https://www.linkedin.com/in/sara-nilsson-example',
            'Talent Acquisition Partner',
            'Sent application follow-up by email.'
        ),
        (
            'Ericsson',
            'Daniel Eriksson',
            'daniel.eriksson@example.com',
            'https://www.linkedin.com/in/daniel-eriksson-example',
            'Backend Developer',
            'Potential technical mentor for the internship.'
        )
) AS seed(
    company_name,
    contact_name,
    email,
    linkedin_url,
    contact_role,
    notes
)
JOIN companies c
    ON c.name = seed.company_name
WHERE NOT EXISTS (
    SELECT 1
    FROM contacts existing
    WHERE existing.company_id = c.id
      AND existing.name = seed.contact_name
      AND existing.email IS NOT DISTINCT FROM seed.email
);

-- ---------------------------------------------------------
-- Application notes
-- ---------------------------------------------------------

INSERT INTO notes (
    application_id,
    content
)
SELECT
    a.id,
    seed.content
FROM (
    VALUES
        (
            'Spotify',
            'Backend Engineering Intern',
            'Application submitted through the company careers page.'
        ),
        (
            'Spotify',
            'Backend Engineering Intern',
            'Review Go concurrency, REST API design, and PostgreSQL before a possible interview.'
        ),
        (
            'Volvo Cars',
            'Software Developer Intern',
            'First interview completed. The team mainly works with Go, Docker, and AWS.'
        ),
        (
            'Klarna',
            'Go Backend Intern',
            'Application was rejected. Ask whether feedback is available.'
        ),
        (
            'Ericsson',
            'Cloud Software Intern',
            'Received an internship offer. Need to review the start date and contract.'
        ),
        (
            'Northvolt',
            'Platform Engineering Intern',
            'Update the CV before submitting the application.'
        )
) AS seed(company_name, application_role, content)
JOIN companies c
    ON c.name = seed.company_name
JOIN applications a
    ON a.company_id = c.id
   AND a.role = seed.application_role
WHERE NOT EXISTS (
    SELECT 1
    FROM notes existing
    WHERE existing.application_id = a.id
      AND existing.content = seed.content
);

-- ---------------------------------------------------------
-- Follow-ups
-- ---------------------------------------------------------

INSERT INTO followups (
    application_id,
    due_date,
    message,
    completed_at
)
SELECT
    a.id,
    seed.due_date,
    seed.message,
    seed.completed_at
FROM (
    VALUES
        (
            'Spotify',
            'Backend Engineering Intern',
            DATE '2026-08-10',
            'Send a follow-up email asking about the application status.',
            NULL::TIMESTAMPTZ
        ),
        (
            'Volvo Cars',
            'Software Developer Intern',
            DATE '2026-08-08',
            'Send a thank-you email after the interview.',
            TIMESTAMPTZ '2026-08-06 14:00:00+02'
        ),
        (
            'Klarna',
            'Go Backend Intern',
            DATE '2026-07-01',
            'Ask the recruiter whether they can provide application feedback.',
            NULL::TIMESTAMPTZ
        ),
        (
            'Ericsson',
            'Cloud Software Intern',
            DATE '2026-08-12',
            'Respond to the offer and confirm the proposed internship dates.',
            NULL::TIMESTAMPTZ
        ),
        (
            'Northvolt',
            'Platform Engineering Intern',
            DATE '2026-08-15',
            'Finish the cover letter and submit the application.',
            NULL::TIMESTAMPTZ
        )
) AS seed(
    company_name,
    application_role,
    due_date,
    message,
    completed_at
)
JOIN companies c
    ON c.name = seed.company_name
JOIN applications a
    ON a.company_id = c.id
   AND a.role = seed.application_role
WHERE NOT EXISTS (
    SELECT 1
    FROM followups existing
    WHERE existing.application_id = a.id
      AND existing.due_date = seed.due_date
      AND existing.message = seed.message
);

COMMIT;
