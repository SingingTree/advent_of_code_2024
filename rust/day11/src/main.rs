
use std::collections::HashMap;
use std::io::BufRead;
use rayon::prelude::*;

fn read_problem_input() -> Vec<u64> {
    let file = std::fs::read("./input").unwrap();
    let lines = file.lines();
    let mut nums = Vec::<u64>::new();
    for line in lines {
        line.unwrap().split_whitespace().map(|x| x.parse().unwrap()).for_each(|x: u64| nums.push(x));
    }
    nums
}

fn get_next_stones(stone: u64) -> Vec<u64> {
    if stone == 0 {
        return vec![1];
    }
    // Check for even length case.
    let as_string = stone.to_string();
    if as_string.len() % 2 == 0 {
        // Even length case.
        let half = as_string.len() / 2;
        let lhs = as_string.chars().take(half).collect::<String>();
        let rhs = as_string.chars().skip(half).take(half).collect::<String>();
        let lhs_num = lhs.parse::<u64>().unwrap();
        let rhs_num = rhs.parse::<u64>().unwrap();
        return vec![lhs_num, rhs_num];
    }
    // Default case.
    vec![
        stone.checked_mul(2024).unwrap()
    ]
}


fn calc_stones(
    stones: Vec<u64>,
    depth_left: u64,
    stone_to_max_depth: &mut HashMap<u64, u64>,
    stone_to_next_stones: &mut HashMap<u64, Vec<u64>>,
) {
    assert!(depth_left > 0);


    let mut frontier = stones;
    let mut i = depth_left;
    while i > 0 {
        let mut new_frontier = Vec::new();
        for &stone in &frontier {
            if stone_to_max_depth.contains_key(&stone) {
                assert!(
                    stone_to_max_depth[&stone] >= depth_left,
                    "Breadth first search means we should neve find a stone with depth_left more than when we first see it",
                );
                if stone_to_max_depth[&stone] > depth_left {
                    // We don't need to check a stone we've already seen.
                    continue;
                }
            } else {
                stone_to_max_depth.insert(stone, depth_left);
            }


            let next_stones = get_next_stones(
                stone,
            );
            stone_to_next_stones.insert(stone, next_stones.clone());
            new_frontier.extend(next_stones);
        }
        new_frontier.sort_unstable();
        new_frontier.dedup();
        frontier = new_frontier;
        i -= 1;
    }
}

fn child_count(
    stone: u64,
    depth_left: u64,
    stone_to_next_stones: &mut HashMap<u64, Vec<u64>>,
    stone_and_depth_to_child_count: &mut HashMap<(u64, u64), u64>,
) -> u64 {
    if depth_left == 1 {
        return stone_to_next_stones[&stone].len() as u64;
    }
    if stone_and_depth_to_child_count.contains_key(&(stone, depth_left)) {
        return stone_and_depth_to_child_count[&(stone, depth_left)];
    }
    let mut count = 0;
    for stone in stone_to_next_stones[&stone].clone() {
        count += child_count(
            stone,
            depth_left - 1,
            stone_to_next_stones,
            stone_and_depth_to_child_count,
        );
    }
    stone_and_depth_to_child_count.insert((stone, depth_left), count);
    count
}

fn main() {
    // Part 1.

    let stones: Vec<u64> = read_problem_input();

    let mut stone_to_max_depth = HashMap::new();
    let mut stone_to_next_stones = HashMap::new();
    calc_stones(stones.clone(), 75, &mut stone_to_max_depth, &mut stone_to_next_stones);

    let mut stone_and_depth_to_child_count = HashMap::new();


    let mut sum = 0u64;
    for &start_stone in &stones {
        sum += child_count(
            start_stone,
            25,
            &mut stone_to_next_stones,
            &mut stone_and_depth_to_child_count,
        );
    }
    println!("{sum}");

    // Part 2.

    let mut sum = 0u64;
    for &start_stone in &stones {
        sum += child_count(
            start_stone,
            75,
            &mut stone_to_next_stones,
            &mut stone_and_depth_to_child_count,
        );
    }
    println!("{sum}");
}
