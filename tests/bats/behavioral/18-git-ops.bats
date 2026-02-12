#!/usr/bin/env bats

load '../lib/test_helper'
load '../lib/assertions'

setup() {
    setup_test_env

    # Initialize PLONK_DIR as a git repo
    git -C "$PLONK_DIR" init -b main
    git -C "$PLONK_DIR" config user.email "test@test.com"
    git -C "$PLONK_DIR" config user.name "Test"
    git -C "$PLONK_DIR" commit --allow-empty -m "initial"
}

@test "add auto-commits to git" {
    local testfile=".plonk-test-rc"
    require_safe_dotfile "$testfile"
    create_test_dotfile "$testfile" "# test"

    run plonk add "$HOME/$testfile"
    assert_success

    # Verify git log has a commit from plonk
    run git -C "$PLONK_DIR" log --oneline -1
    assert_output --partial "plonk: add"
}

@test "rm auto-commits to git" {
    local testfile=".plonk-test-rc"
    require_safe_dotfile "$testfile"
    create_test_dotfile "$testfile" "# test"

    # Add first
    plonk add "$HOME/$testfile"

    # Remove
    run plonk rm "$HOME/$testfile"
    assert_success

    run git -C "$PLONK_DIR" log --oneline -1
    assert_output --partial "plonk: rm"
}

@test "auto-commit disabled via config" {
    echo 'git:' > "$PLONK_DIR/plonk.yaml"
    echo '  auto_commit: false' >> "$PLONK_DIR/plonk.yaml"
    git -C "$PLONK_DIR" add -A && git -C "$PLONK_DIR" commit -m "add config"

    local testfile=".plonk-test-rc"
    require_safe_dotfile "$testfile"
    create_test_dotfile "$testfile" "# test"

    run plonk add "$HOME/$testfile"
    assert_success

    # Verify repo is dirty (not auto-committed)
    run git -C "$PLONK_DIR" status --porcelain
    refute_output ""
}

@test "non-git-repo warns about git" {
    # Remove .git from PLONK_DIR
    rm -rf "$PLONK_DIR/.git"

    local testfile=".plonk-test-rc"
    require_safe_dotfile "$testfile"
    create_test_dotfile "$testfile" "# test"

    run plonk add "$HOME/$testfile"
    assert_success
    assert_output --partial "not a git repository"
    assert_output --partial "auto_commit"
}

@test "push command works with remote" {
    # Set up a bare remote
    local remote_dir="$BATS_TEST_TMPDIR/remote.git"
    git init --bare -b main "$remote_dir"
    git -C "$PLONK_DIR" remote add origin "$remote_dir"
    git -C "$PLONK_DIR" push -u origin main

    # Create a change
    echo "test" > "$PLONK_DIR/testfile"
    git -C "$PLONK_DIR" add -A
    git -C "$PLONK_DIR" commit -m "test change"

    run plonk push
    assert_success
    assert_output --partial "Push complete"
}

@test "push warns about uncommitted changes" {
    local remote_dir="$BATS_TEST_TMPDIR/remote.git"
    git init --bare -b main "$remote_dir"
    git -C "$PLONK_DIR" remote add origin "$remote_dir"
    git -C "$PLONK_DIR" push -u origin main

    # Create an uncommitted change
    echo "uncommitted" > "$PLONK_DIR/newfile"

    run plonk push
    assert_success
    assert_output --partial "Warning: uncommitted changes"
    assert_output --partial "Push complete"

    # Verify the file is still uncommitted
    run git -C "$PLONK_DIR" status --porcelain
    assert_output --partial "newfile"
}

@test "push fails without remote" {
    run plonk push
    assert_failure
    assert_output --partial "no remote configured"
}

@test "push fails for non-git-repo" {
    rm -rf "$PLONK_DIR/.git"

    run plonk push
    assert_failure
    assert_output --partial "not a git repository"
}

@test "pull command works" {
    local remote_dir="$BATS_TEST_TMPDIR/remote.git"
    git init --bare -b main "$remote_dir"
    git -C "$PLONK_DIR" remote add origin "$remote_dir"
    git -C "$PLONK_DIR" push -u origin main

    # Create a commit in a clone and push it
    local clone_dir="$BATS_TEST_TMPDIR/clone"
    git clone "$remote_dir" "$clone_dir"
    git -C "$clone_dir" config user.email "test@test.com"
    git -C "$clone_dir" config user.name "Test"
    echo "remote change" > "$clone_dir/remotefile"
    git -C "$clone_dir" add -A
    git -C "$clone_dir" commit -m "remote commit"
    git -C "$clone_dir" push

    run plonk pull
    assert_success
    assert_output --partial "Pull complete"

    # Verify the file arrived
    [ -f "$PLONK_DIR/remotefile" ]
}

@test "pull auto-commits dirty state" {
    local remote_dir="$BATS_TEST_TMPDIR/remote.git"
    git init --bare -b main "$remote_dir"
    git -C "$PLONK_DIR" remote add origin "$remote_dir"
    git -C "$PLONK_DIR" push -u origin main

    # Create local uncommitted change
    echo "local" > "$PLONK_DIR/localfile"

    run plonk pull
    assert_success
    assert_output --partial "Committed local changes before pull"

    # Verify file was committed
    run git -C "$PLONK_DIR" status --porcelain
    assert_output ""
}

@test "pull --apply runs apply after pull" {
    local remote_dir="$BATS_TEST_TMPDIR/remote.git"
    git init --bare -b main "$remote_dir"
    git -C "$PLONK_DIR" remote add origin "$remote_dir"
    git -C "$PLONK_DIR" push -u origin main

    run plonk pull --apply
    assert_success
    assert_output --partial "Pull complete"
    assert_output --partial "Applying configuration"
}

@test "pull fails for non-git-repo" {
    rm -rf "$PLONK_DIR/.git"

    run plonk pull
    assert_failure
    assert_output --partial "not a git repository"
}

@test "pull fails without remote" {
    run plonk pull
    assert_failure
    assert_output --partial "no remote configured"
}

@test "pull refuses dirty state when auto_commit disabled" {
    local remote_dir="$BATS_TEST_TMPDIR/remote.git"
    git init --bare -b main "$remote_dir"
    git -C "$PLONK_DIR" remote add origin "$remote_dir"
    git -C "$PLONK_DIR" push -u origin main

    # Disable auto_commit and commit the config change
    echo 'git:' > "$PLONK_DIR/plonk.yaml"
    echo '  auto_commit: false' >> "$PLONK_DIR/plonk.yaml"
    git -C "$PLONK_DIR" add -A && git -C "$PLONK_DIR" commit -m "disable auto_commit"
    git -C "$PLONK_DIR" push

    # Now create a separate dirty file
    echo "dirty" > "$PLONK_DIR/uncommitted-file"

    run plonk pull
    assert_failure
    assert_output --partial "uncommitted changes"
}
