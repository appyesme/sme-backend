begin;

    alter database postgres set timezone to 'Asia/Kolkata';

    create table if not exists phone_verifications (
        id uuid primary key not null default uuid_generate_v4(),
        created_at timestamp with time zone default current_timestamp not null,
        updated_at timestamp with time zone default current_timestamp not null,
        phone_number varchar not null,
        otp_code varchar(6) not null,
        otp_expires_at timestamp not null,

        unique (phone_number)
    );

    create table if not exists auth (
        id uuid primary key not null default uuid_generate_v4(),
        created_at timestamp with time zone default current_timestamp not null,
        updated_at timestamp with time zone default current_timestamp not null,
        phone_number varchar not null,
        user_type text not null,
        verified_at timestamp with time zone default null,

        unique (phone_number)
    );

    create table if not exists users (
        id uuid primary key not null references auth(id) on delete cascade,
        created_at timestamp with time zone default current_timestamp not null,
        updated_at timestamp with time zone default current_timestamp not null,
        name text not null,
        email text default null,
        photo_url text default null,
        expertises text[] default null,
        total_work_experience smallint default 0,
        documents text[] default null,
        aadhar_number varchar(12) default null,
        pan_number varchar(10) default null,
        about text default null,
        verified boolean default false
    );

    create table if not exists favourite_users (
        id uuid not null primary key default uuid_generate_v4(),
        created_at timestamp with time zone default current_timestamp not null,
        updated_at timestamp with time zone default current_timestamp not null,
        created_by uuid not null references users(id) on delete cascade,
        profile_id uuid not null references users(id) on delete cascade,
        unique(profile_id, created_by)
    );

    create table if not exists services (
        id uuid not null primary key default uuid_generate_v4(),
        created_at timestamp with time zone default current_timestamp not null,
        updated_at timestamp with time zone default current_timestamp not null,
        created_by uuid not null references users(id) on delete cascade,
        title text not null,
        expertises text not null,
        charge float not null,
        additional_charge float not null default 0.0,
        home_available boolean not null default false,
        salon_available boolean not null default false,
        description text default null,
        address text default null,
        status text not null default 'DRAFTED' -- PUBLISHED, DRAFTED
    );

    create table if not exists service_medias (
        id uuid not null primary key default uuid_generate_v4(),
        created_at timestamp with time zone default current_timestamp not null,
        updated_at timestamp with time zone default current_timestamp not null,
        created_by uuid not null references users(id) on delete cascade,
        service_id uuid not null references services(id) on delete cascade,
        file_name text not null,
        url text not null,
        storage_path text not null
    );

    create table if not exists service_days (
        id uuid not null primary key default uuid_generate_v4(),
        created_at timestamp with time zone default current_timestamp not null,
        updated_at timestamp with time zone default current_timestamp not null,
        created_by uuid not null references users(id) on delete cascade,
        service_id uuid not null references services(id) on delete cascade,
        day smallint not null, -- SUN(0) - SAT(6)
        enabled boolean not null default false,

        unique(created_by, service_id, day)
    );

    create table if not exists service_timings (
        id uuid not null primary key default uuid_generate_v4(),
        created_at timestamp with time zone default current_timestamp not null,
        updated_at timestamp with time zone default current_timestamp not null,
        created_by uuid not null references users(id) on delete cascade,
        service_day_id uuid not null references service_days(id) on delete cascade,
        start_time time not null,
        end_time time not null,
        people_per_slot smallint not null,
        enabled boolean not null default false
    );

    create table if not exists appointments (
        id uuid not null primary key default uuid_generate_v4(),
        created_at timestamp with time zone default current_timestamp not null,
        updated_at timestamp with time zone default current_timestamp not null,
        created_by uuid not null references users(id),
        service_id uuid not null references services(id),
        service_timing_id uuid not null references service_timings(id),
        appointment_date date not null,
        home_service_needed boolean not null default false,
        start_time time not null,
        end_time time not null,
        status text not null default 'INITIATED' -- INITIATED, BOOKED, CANCELLED, ONGOING, COMPLETED
    );

    create table if not exists payments (
        id uuid not null primary key default uuid_generate_v4(),
        created_at timestamp with time zone default current_timestamp not null,
        updated_at timestamp with time zone default current_timestamp not null,
        created_by uuid not null references users(id),
        service_id uuid not null references services(id),
        appointment_id uuid not null references appointments(id),
        amount numeric(10, 2) not null,
        currency varchar(10) default 'INR',
        order_id varchar(255),
        payment_id varchar(255),
        status varchar(50) not null, -- PENDING, PAID, CANCELLED, REFUNDED

        unique(appointment_id, created_by)
    );

    create table if not exists last_payment_cleared (
        id uuid not null primary key default uuid_generate_v4(),
        created_at timestamp with time zone default current_timestamp not null,
        updated_at timestamp with time zone default current_timestamp not null,
        created_by uuid not null references users(id) on delete cascade,
        entrepreneur_id uuid not null references users(id) on delete cascade,
        unique(entrepreneur_id)
    );

    create table if not exists refunds (
        id uuid not null primary key default uuid_generate_v4(),
        created_at timestamp with time zone default current_timestamp not null,
        updated_at timestamp with time zone default current_timestamp not null,
        payment_id uuid not null references payments(id),
        refunded_at timestamp with time zone default current_timestamp not null,
        refund_amount numeric(10, 2) NOT NULL,
        refund_status varchar(50) DEFAULT 'PENDING', -- PENDING, COMPLETED, FAILED
        refund_reason varchar(255)
    );

    create table if not exists posts (
        id uuid not null primary key default uuid_generate_v4(),
        created_at timestamp with time zone default current_timestamp not null,
        updated_at timestamp with time zone default current_timestamp not null,
        created_by uuid not null references users(id) on delete cascade,
        service_id uuid not null references services(id) on delete cascade,
        description text not null,
        status text not null default 'DRAFTED' -- PUBLISHED, DRAFTED
    );

    create table if not exists post_medias (
        id uuid not null primary key default uuid_generate_v4(),
        created_at timestamp with time zone default current_timestamp not null,
        updated_at timestamp with time zone default current_timestamp not null,
        post_id uuid not null references posts(id) on delete cascade,
        created_by uuid not null references users(id) on delete cascade,
        file_name text not null,
        url text not null,
        storage_path text not null
    );

    create table if not exists fcm_tokens (
        id uuid primary key not null default uuid_generate_v4(),
        created_at timestamp with time zone default current_timestamp not null,
        created_by uuid not null references users (id) on delete cascade,
        token text not null,
        unique(created_by, token)
    );

    create table if not exists notifications (
        id uuid primary key not null default uuid_generate_v4(),
        created_at timestamp with time zone default current_timestamp not null,
        updated_at timestamp with time zone default current_timestamp not null,
        user_id uuid not null references users (id) on delete cascade,
        actions jsonb not null default '[]'::jsonb,
        title varchar not null default '',
        body text not null default '',
        read boolean not null default false,
        visible boolean not null default true,
        fcm_status varchar not null default 'pending'
    );

    create table if not exists bank_accounts (
        id uuid primary key not null default uuid_generate_v4(),
        created_at timestamp with time zone default current_timestamp not null,
        updated_at timestamp with time zone default current_timestamp not null,
        created_by uuid not null references users (id) on delete cascade,
        account_name text not null,
        account_number varchar not null,
        ifsc_code varchar not null,
        upi varchar,

        unique(created_by)
    );

commit;
