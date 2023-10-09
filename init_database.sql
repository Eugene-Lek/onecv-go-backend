CREATE TABLE IF NOT EXISTS teacher (
    email TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS student (
    email TEXT PRIMARY KEY,
    suspended BOOLEAN
);

CREATE TABLE IF NOT EXISTS teacher_student_relationship (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    teacher TEXT,
    student TEXT,

    UNIQUE(teacher, student),

    CONSTRAINT fk_teacher
        FOREIGN KEY (teacher)
            REFERENCES teacher(email)
            ON UPDATE CASCADE
            ON DELETE CASCADE,

    CONSTRAINT fk_student
        FOREIGN KEY (student)
            REFERENCES student(email)
            ON UPDATE CASCADE
            ON DELETE CASCADE
);